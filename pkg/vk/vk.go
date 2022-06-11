package vk

import (
	"errors"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "vk",
}

func GetUserInfo(vkapi *api.VK, userID string) (*object.UsersUser, error) {
	users, err := vkapi.UsersGet(api.Params{
		"user_ids": userID,
		"fields":   "screen_name",
	})

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New(fmt.Sprintf("User with id %d not found", userID))
	}

	return &users[0], err
}
