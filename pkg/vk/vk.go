package vk

import (
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/SevereCloud/vksdk/v2/object"
)

func GetUserInfo(vkapi *api.VK, event wrapper.ChatInfoChange) (*object.UsersUser, error) {
	users, err := vkapi.UsersGet(api.Params{
		"user_ids": event.Info,
		"fields":   "screen_name",
	})

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New(fmt.Sprintf("User with id %d not found", event.Info))
	}

	return &users[0], err
}
