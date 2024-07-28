package main

import "github.com/Tsundere-Musume/message/internal/validator"

type SignUpForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type LogInForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type MessageForm struct {
	Message   string `form:"message"`
	CSRFToken string `form:"csrf_token"`
}
