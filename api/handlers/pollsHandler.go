package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
)

const (
	reddit_uid_test = "sqv1rf88"
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
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test
	log.Info("reddit_uid: ", reddit_uid)

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	input := data.Filters{}

	qs := c.Request().URL.Query()
	input.Page = h.Utils.ReadIntQuery(qs, "page", 1)
	input.PageSize = h.Utils.ReadIntQuery(qs, "page_size", 10)

	err = h.Validate.Struct(input)
	if err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	if reddit_uid == "" {
		polls, metadata, err := h.Data.Polls.GetAllPolls(sub, input)
		if err != nil {
			h.Utils.InternalServerError(c, fmt.Errorf("error in getting polls; %v", err))
			return err
		}
		if len(polls) == 0 {
			h.Utils.CustomErrorResponse(c, utils.Cake{"error": "no polls found"}, http.StatusNotFound, nil)
			return nil
		}

		return c.JSON(http.StatusOK, Cake{"polls": polls, "metadata": metadata})
	}

	polls, metadata, err := h.Data.Polls.GetAllPollsSigned(reddit_uid, sub, input)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in getting polls; %v", err))
		return err
	}
	// if len(polls) == 0 {
	// 	h.Utils.CustomErrorResponse(c, utils.Cake{"error": "no polls found"}, http.StatusNotFound, nil)
	// 	return nil
	// }

	return c.JSON(http.StatusOK, Cake{"polls": polls, "metadata": metadata})

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

	rows, err := h.Data.Polls.CreatePollVote(pollID, reddit_uid, int(optionID))
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in inserting poll vote; %v", err))
		return err
	}

	if rows == 0 && err == nil {
		TLerror := fmt.Errorf("update time limit Exceeded")
		h.Utils.CustomErrorResponse(c, utils.Cake{"message": "you've exceeded time limit to make change"}, http.StatusAlreadyReported, TLerror)
		return TLerror
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
