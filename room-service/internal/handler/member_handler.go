package handler

import "github.com/FelipeFelipeRenan/goverse/room-service/internal/service"

type MemberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(memberService service.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}
