package vk

import (
	"errors"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
)

var logPackage = "vk"

type MediaAttachmentType string

const (
	AudioType    MediaAttachmentType = "audio"
	PhotoType    MediaAttachmentType = "photo"
	VideoType    MediaAttachmentType = "video"
	DocumentType MediaAttachmentType = "doc"
)

func GetUserInfo(vkapi *api.VK, userID string) (*object.UsersUser, error) {
	users, err := vkapi.UsersGet(api.Params{
		"user_ids": userID,
		"fields":   "screen_name",
	})

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New(fmt.Sprintf("user with id `%s` not found", userID))
	}

	return &users[0], err
}

func GetWallPostsCount(vkapi *api.VK, community string) (int, error) {
	response, err := vkapi.WallGet(api.Params{
		"domain": community,
		"count":  1,
	})

	if err != nil {
		return 0, err
	}

	return response.Count, nil
}
