package socketio_service

import (
	"encoding/json"
	"log"
	"what-to-eat/be/model"
	"what-to-eat/be/service"
)

func ProcessDishVoteUpdate(data ...any) (*model.DishVote, model.SocketioJoinRoom) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Error marshaling data[0] to JSON: %v", err)
	}
	var input model.SocketioDishVoteUpdate
	if err := json.Unmarshal(jsonStr, &input); err != nil {
		log.Printf("Error unmarshaling JSON to SocketioDishVoteUpdate: %v", err)
	}

	jsonStrRoom, err := json.Marshal(data[1])
	if err != nil {
		log.Printf("Error marshaling data[1] to JSON: %v", err)
	}
	var socketioJoinRoomData model.SocketioJoinRoom
	if err := json.Unmarshal(jsonStrRoom, &socketioJoinRoomData); err != nil {
		log.Printf("Error unmarshaling JSON to SocketioJoinRoom: %v", err)
	}

	var dishVoteService service.DishVoteService

	dishVote, err := dishVoteService.FindOne(socketioJoinRoomData.RoomID)
	if err != nil {
		log.Printf("Error finding DishVote with RoomID %s: %v", socketioJoinRoomData.RoomID, err)
	}

	updateDishVote := model.UpdateDishVoteDto{
		ID:            dishVote.ID,
		Title:         dishVote.Title,
		Description:   dishVote.Description,
		DishVoteItems: []*model.DishVoteItem{},
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
			if index >= 0 {
				dishVote.DishVoteItems[i].VoteAnonymous = append(dishVote.DishVoteItems[i].VoteAnonymous[:index], dishVote.DishVoteItems[i].VoteAnonymous[index+1:]...)
			}
		}

		item := model.DishVoteItem{
			Slug:          dishVote.DishVoteItems[i].Slug,
			VoteUser:      dishVote.DishVoteItems[i].VoteUser,
			VoteAnonymous: dishVote.DishVoteItems[i].VoteAnonymous,
			IsCustom:      dishVote.DishVoteItems[i].IsCustom,
			CustomTitle:   dishVote.DishVoteItems[i].CustomTitle,
		}
		updateDishVote.DishVoteItems = append(updateDishVote.DishVoteItems, &item)
	}

	updated, err := dishVoteService.Update(updateDishVote, nil)

	if err != nil {
		log.Printf("Error updating dish vote: %v", err)
	}

	return updated, socketioJoinRoomData
}
