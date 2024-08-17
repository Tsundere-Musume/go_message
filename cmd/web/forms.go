package main

import (
	"github.com/Tsundere-Musume/message/internal/validator"
)

type SignUpForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	Avatar              []byte `form:"avatar"`
	validator.Validator `form:"-"`
}

type LogInForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type DirectMessageForm struct {
	Message    string `form:"message"`
	SenderID   string `form:"senderId"`
	ReceiverID string `form:"receiverId"`
	CSRFToken  string `form:"csrf_token"`
	Timezone   string `form:"timezone"`
	//TODO: maybe validation to check if the sender can send direct methods to the receiver
}
