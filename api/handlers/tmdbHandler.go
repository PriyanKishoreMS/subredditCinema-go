package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type response struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func (h *Handlers) SearchActorsHandler(c echo.Context) error {
	name, err := h.Utils.ReadStringParam(c, "name")
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching hero name: %v", err))
		return c.JSON(http.StatusBadRequest, Cake{
			"message": "Invalid hero name",
			"status":  "error",
		})
	}

	page := h.Utils.ReadStringQuery(c.QueryParams(), "page", "1")
	pageInt, _ := strconv.Atoi(page)
	if pageInt < 1 {
		return c.JSON(http.StatusBadRequest, Cake{
			"message": "Invalid page number",
			"status":  "error",
		})
	}

	options := map[string]string{
		"append_to_response": "images",
		"page":               page,
	}
	width := 300

	actors, err := h.Tmdb.GetSearchPeople(name, options)
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching actor details: %v", err))
		return c.JSON(http.StatusInternalServerError, Cake{
			"message": "Error fetching actor details",
			"status":  "error",
		})
	}

	res := []response{}

	if actors == nil || actors.Results == nil {
		return c.JSON(http.StatusNotFound, Cake{
			"message": "No people found",
			"status":  "error",
		},
		)
	}

	for _, v := range actors.Results {
		if v.ProfilePath != "" {
			res = append(res, response{
				Id:    v.ID,
				Name:  v.Name,
				Image: fmt.Sprintf("https://image.tmdb.org/t/p/w%d%s", width, v.ProfilePath),
			})
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handlers) SearchMoviesHandler(c echo.Context) error {
	name, err := h.Utils.ReadStringParam(c, "name")
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching movie name: %v", err))
		return c.JSON(http.StatusBadRequest, Cake{
			"message": "Invalid movie name",
			"status":  "error",
		})
	}

	page := h.Utils.ReadStringQuery(c.QueryParams(), "page", "1")
	pageInt, _ := strconv.Atoi(page)

	if pageInt < 1 {
		return c.JSON(http.StatusBadRequest, Cake{
			"message": "Invalid page number",
			"status":  "error",
		})

	}

	options := map[string]string{
		"append_to_response": "images",
		"page":               page,
	}

	width := 300

	movies, err := h.Tmdb.GetSearchMovies(name, options)
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching movie details: %v", err))
		return c.JSON(http.StatusInternalServerError, Cake{
			"message": "Error fetching movie details",
			"status":  "error",
		})
	}

	res := []response{}

	if movies == nil || movies.Results == nil {
		return c.JSON(http.StatusNotFound, Cake{
			"message": "No movies found",
			"status":  "error",
		},
		)
	}
	for _, v := range movies.Results {
		if v.PosterPath != "" {
			res = append(res, response{
				Id:    v.ID,
				Name:  v.Title,
				Image: fmt.Sprintf("https://image.tmdb.org/t/p/w%d/%s", width, v.PosterPath),
			})
		}
	}

	return c.JSON(http.StatusOK, res)
}
