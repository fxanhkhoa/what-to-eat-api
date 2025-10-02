package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeedbackController struct{}

func NewFeedbackController() *FeedbackController {
	return &FeedbackController{}
}

// Create - Public endpoint to create feedback (no authentication required)
func (fc *FeedbackController) Create(c echo.Context) error {
	var dto model.CreateFeedbackDto
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data",
		})
	}

	// Validate the DTO
	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Get user agent from request
	dto.UserAgent = c.Request().UserAgent()

	// Check if user is authenticated (optional)
	var userId *primitive.ObjectID
	claim := c.Get("CLAIM")
	if claim != nil {
		userClaim := claim.(*model.JwtCustomClaims)
		if userClaim.ID != "" {
			objectId, err := primitive.ObjectIDFromHex(userClaim.ID)
			if err == nil {
				userId = &objectId
			}
		}
	}

	feedbackService := &service.FeedbackService{}
	feedback, err := feedbackService.Create(dto, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create feedback",
		})
	}

	return c.JSON(http.StatusCreated, feedback)
}

// GetAll - Get all feedbacks with pagination (requires authentication)
func (fc *FeedbackController) GetAll(c echo.Context) error {
	var query model.FeedbackListDto

	var err error
	query.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.Page < 0 {
		query.Page = 1
	}

	query.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.Limit < 0 {
		query.Limit = 10
	}

	// Optional filters
	if ratingStr := c.QueryParam("rating"); ratingStr != "" {
		rating, err := strconv.Atoi(ratingStr)
		if err == nil && rating >= 1 && rating <= 5 {
			query.Rating = &rating
		}
	}

	if email := c.QueryParam("email"); email != "" {
		query.Email = &email
	}

	feedbackService := &service.FeedbackService{}
	result, err := feedbackService.GetAll(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve feedbacks",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetById - Get a specific feedback by ID (requires authentication)
func (fc *FeedbackController) GetById(c echo.Context) error {
	id := c.Param("id")

	feedbackService := &service.FeedbackService{}
	feedback, err := feedbackService.GetById(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Feedback not found",
		})
	}

	return c.JSON(http.StatusOK, feedback)
}

// Update - Update user's own feedback (requires authentication)
func (fc *FeedbackController) Update(c echo.Context) error {
	id := c.Param("id")

	var dto model.UpdateFeedbackDto
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data",
		})
	}

	// Validate the DTO
	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	claim := c.Get("CLAIM").(*model.JwtCustomClaims)

	userObjectId, err := primitive.ObjectIDFromHex(claim.ID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
		})
	}

	feedbackService := &service.FeedbackService{}
	feedback, err := feedbackService.Update(id, dto, userObjectId)
	if err != nil {
		if err.Error() == "unauthorized to update this feedback" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update feedback",
		})
	}

	return c.JSON(http.StatusOK, feedback)
}

// Delete - Delete user's own feedback (requires authentication)
func (fc *FeedbackController) Delete(c echo.Context) error {
	id := c.Param("id")

	claim := c.Get("CLAIM").(*model.JwtCustomClaims)

	userObjectId, err := primitive.ObjectIDFromHex(claim.ID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
		})
	}

	feedbackService := &service.FeedbackService{}
	err = feedbackService.Delete(id, userObjectId)
	if err != nil {
		if err.Error() == "unauthorized to delete this feedback" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete feedback",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Feedback deleted successfully",
	})
}
