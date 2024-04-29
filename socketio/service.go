package socketio_helper

import (
	"encoding/json"
	"fmt"
	"time"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"
)

func ProcessDishVoteUpdate(data ...any) (*model.DishVote, SocketioJoinRoom) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		fmt.Println(err)
	}
	var input SocketioDishVoteUpdate
	if err := json.Unmarshal(jsonStr, &input); err != nil {
		fmt.Println(err)
	}

	jsonStrRoom, err := json.Marshal(data[1])
	if err != nil {
		fmt.Println(err)
	}
	var socketioJoinRoomData SocketioJoinRoom
	if err := json.Unmarshal(jsonStrRoom, &socketioJoinRoomData); err != nil {
		fmt.Println(err)
	}

	dishVote, err := service.NewDishVoteService().FindOne(socketioJoinRoomData.RoomID)
	if err != nil {
		fmt.Println(err.Error())
	}

	updateDishVote := model.UpdateDishVoteInput{
		ID:            dishVote.ID,
		Title:         dishVote.Title,
		Description:   dishVote.Description,
		DishVoteItems: []*model.DishVoteItemInput{},
	}

	for i, val := range dishVote.DishVoteItems {
		if val.Slug == input.Slug && input.IsVoting {
			dishVote.DishVoteItems[i].VoteAnonymous = append(dishVote.DishVoteItems[i].VoteAnonymous, &input.MyName)
		} else if val.Slug == input.Slug && !input.IsVoting {
			index := -1
			for j, vote := range dishVote.DishVoteItems[i].VoteAnonymous {
				if *vote == input.MyName {
					index = j
					break
				}
			}
			dishVote.DishVoteItems[i].VoteAnonymous = append(dishVote.DishVoteItems[i].VoteAnonymous[:index], dishVote.DishVoteItems[i].VoteAnonymous[index+1:]...)
		}

		item := model.DishVoteItemInput{
			Slug:          dishVote.DishVoteItems[i].Slug,
			VoteUser:      dishVote.DishVoteItems[i].VoteUser,
			VoteAnonymous: dishVote.DishVoteItems[i].VoteAnonymous,
			IsCustom:      dishVote.DishVoteItems[i].IsCustom,
		}
		updateDishVote.DishVoteItems = append(updateDishVote.DishVoteItems, &item)
	}
	user := &model.User{
		Email:       "",
		Password:    new(string),
		Name:        new(string),
		DateOfBirth: &time.Time{},
		Address:     new(string),
		Phone:       new(string),
		GoogleID:    new(string),
		FacebookID:  new(string),
		GithubID:    new(string),
		Avatar:      new(string),
		Deleted:     false,
		DeletedAt:   &time.Time{},
		DeletedBy:   new(string),
		UpdatedAt:   &time.Time{},
		UpdatedBy:   new(string),
		CreatedAt:   &time.Time{},
		CreatedBy:   new(string),
		ID:          "",
		RoleName:    "",
	}
	updated, err := service.NewDishVoteService().Update(updateDishVote, user)

	if err != nil {
		fmt.Println(err.Error())
	}

	return updated, socketioJoinRoomData
}
