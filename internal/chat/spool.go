package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// spool é uma fila durável em arquivo (JSON por linha) para mensagens que não puderam ir ao
// Postgres. A mensagem já tem ID (UUIDv7) e horário definidos, então o reprocessamento é idempotente
// (INSERT ... ON CONFLICT DO NOTHING) e a ordem no histórico independe de quando for gravada.
//
// Garantia: sobrevive a restart do processo, desde que o diretório sobreviva. Não protege contra
// perda do disco. Uma mensagem aceita no spool já foi entregue ao vivo, mas só aparece no histórico
// depois do reprocessamento.
type spool struct {
	mu   sync.Mutex
	path string
	pend atomic.Int64
}

func newSpool(dir string) (*spool, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	s := &spool{path: filepath.Join(dir, "pending.jsonl")}
	if data, err := os.ReadFile(s.path); err == nil {
		s.pend.Store(int64(bytes.Count(data, []byte("\n"))))
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *spool) pending() int64 { return s.pend.Load() }

// append grava e dá fsync antes de retornar: só depois disso a mensagem conta como aceita.
func (s *spool) append(m Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	s.pend.Add(1)
	return nil
}

// drain reprocessa em ordem. Para no primeiro erro transitório e mantém o resto no arquivo.
// Erro permanente vai para dead() e a mensagem sai da fila. Devolve quantas saíram da fila.
func (s *spool) drain(ctx context.Context,
	save func(context.Context, Message) error,
	dead func(context.Context, Message, string) error) (int, error) {

	s.mu.Lock()
	data, err := os.ReadFile(s.path)
	s.mu.Unlock()
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	offset, done := 0, 0
	var firstErr error
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		if ctx.Err() != nil {
			firstErr = ctx.Err()
			break
		}
		var m Message
		if err := json.Unmarshal(bytes.TrimSpace(line), &m); err != nil {
			slog.Error("chat: linha corrompida no spool, descartada", "err", err, "linha", string(line))
		} else if err := save(ctx, m); err != nil {
			if !isPermanent(err) {
				firstErr = err
				break
			}
			if derr := dead(ctx, m, err.Error()); derr != nil {
				firstErr = derr
				break
			}
			slog.Error("chat: mensagem enviada para dead_letters", "id", m.ID, "err", err)
		}
		offset += len(line)
		done++
	}
	if offset == 0 {
		return 0, firstErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	cur, err := os.ReadFile(s.path)
	if err != nil {
		return done, err
	}
	// O arquivo só cresce por append: o prefixo [0:offset] é o que acabamos de processar.
	rest := cur[offset:]
	if len(rest) == 0 {
		err = os.Remove(s.path)
	} else {
		tmp := s.path + ".tmp"
		if err = os.WriteFile(tmp, rest, 0o640); err == nil {
			err = os.Rename(tmp, s.path)
		}
	}
	if err != nil {
		return done, err
	}
	s.pend.Add(int64(-done))
	return done, firstErr
}
