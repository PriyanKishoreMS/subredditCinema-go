package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
)

func (h *Handlers) CreateTierListHandler(c echo.Context) error {
	reddit_uid := reddit_uid_test
	// reddit_uid := c.Get("reddit_uid").(string)

	var input *data.TierListData

	if err := h.Utils.ReadJSON(c, &input); err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading json; %v", err))
		return err
	}

	err := h.Validate.Struct(input)
	if err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	err = h.Data.Tierlists.CreateNewTierListTemplate(reddit_uid, *input)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in creating tier list template; %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"message": "tier list template created"})
}

func (h *Handlers) GetAllTierlistHandler(c echo.Context) error {
	filters := data.Filters{}

	qs := c.Request().URL.Query()
	filters.Page = h.Utils.ReadIntQuery(qs, "page", 1)
	filters.PageSize = h.Utils.ReadIntQuery(qs, "page_size", 10)

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading subreddit; %v", err))
		return err
	}

	tierLists, metadata, err := h.Data.Tierlists.GetAllTierlists(sub, filters)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in getting tier lists; %v", err))
		return err
	}

	return c.JSON(http.StatusOK, Cake{"tier_lists": tierLists, "metadata": metadata})
}

func (h *Handlers) GetTierListByIDHandler(c echo.Context) error {
	tierListID, err := h.Utils.ReadIntParam(c, "id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading tier list id; %v", err))
		return err
	}

	tierList, err := h.Data.Tierlists.GetTierListByID(tierListID)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in getting tier list; %v", err))
		return err
	}

	return c.JSON(http.StatusOK, tierList)
}
