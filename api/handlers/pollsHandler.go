package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
)

const (
	reddit_uid_test = "eh2wrd0r"
)

func (h *Handlers) CreatePollHandler(c echo.Context) error {
	var input *data.Poll

	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test

	if err := h.Utils.ReadJSON(c, &input); err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading json; %v", err))
		return err
	}

	input.RedditUID = reddit_uid

	// todo Uncomment before deploying
	// pollLimit, err := h.Data.Polls.PollLimitForUser(reddit_uid)
	// if err != nil {
	// 	h.Utils.InternalServerError(c, fmt.Errorf("error in getting poll limit for user; %v", err))
	// 	return err
	// }

	// if pollLimit {
	// 	h.Utils.CustomErrorResponse(c, utils.Cake{"error": "user can't create poll now"}, http.StatusBadRequest, nil)
	// 	return nil
	// }

	if input.VotingMethod == "" {
		input.VotingMethod = "reddit_oauth2"
	}

	if input.EndTime.IsZero() {
		input.EndTime = time.Now().Add(24 * time.Hour)
	}

	var options []data.PollOption

	if err := json.Unmarshal(input.Options, &options); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in unmarshalling options; %v", err))
		return err
	}

	if err := h.Validate.Struct(input); err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	for _, option := range options {
		if err := h.Validate.Struct(option); err != nil {
			h.Utils.ValidationError(c, err)
			return err
		}
	}

	if err := h.Data.Polls.InsertNewPoll(input, options); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in inserting poll; %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"message": "poll created"})
}

func (h *Handlers) GetAllPollsHandler(c echo.Context) error {
	polls, err := h.Data.Polls.GetAllPolls()
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in getting polls; %v", err))
		return err
	}
	if len(polls) == 0 {
		h.Utils.CustomErrorResponse(c, utils.Cake{"error": "no polls found"}, http.StatusNotFound, nil)
		return nil
	}

	return c.JSON(http.StatusOK, polls)
}

func (h *Handlers) GetPollByIDHandler(c echo.Context) error {
	pollID, err := h.Utils.ReadIntParam(c, "poll_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading poll_id; %v", err))
		return err
	}

	poll, err := h.Data.Polls.GetPollByID(pollID)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in getting poll; %v", err))
		return err
	}

	return c.JSON(http.StatusOK, poll)
}

func (h *Handlers) CreatePollVoteHandler(c echo.Context) error {

	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test

	pollID, err := h.Utils.ReadIntParam(c, "poll_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading poll_id; %v", err))
		return err
	}

	optionID, err := h.Utils.ReadIntParam(c, "option_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading option_id; %v", err))
		return err
	}

	if err := h.Data.Polls.CreatePollVote(pollID, reddit_uid, int(optionID)); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in inserting poll vote; %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"message": "vote created"})
}

func (h *Handlers) DeletePollByCreatorHandler(c echo.Context) error {
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test

	pollID, err := h.Utils.ReadIntParam(c, "poll_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading poll_id; %v", err))
		return err
	}

	if err := h.Data.Polls.DeletePollByCreator(pollID, reddit_uid); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in deleting poll; %v", err))
	}

	return c.JSON(http.StatusOK, Cake{"message": "poll deleted"})
}
