package router

import "github.com/labstack/echo/v4"

func InitializeRoutes(e *echo.Echo) {
	authGroup := e.Group("/auth")
	UseAuthGroup(authGroup)

	userGroup := e.Group("/user")
	UseUserGroup(userGroup)

	dishGroup := e.Group("/dish")
	UseDishRouter(dishGroup)

	ingredientGroup := e.Group("/ingredient")
	UseIngredientRouter(ingredientGroup)

	dishVoteGroup := e.Group("/dish-vote")
	UseDishVoteRouter(dishVoteGroup)

	rolePermissionGroup := e.Group("/authorization")
	UseRolePermissionRouter(rolePermissionGroup)

	contactGroup := e.Group("/contact")
	UseContactRouter(contactGroup)

	chatGroup := e.Group("/chat")
	UseChatRouter(chatGroup)

	userLoginTrackGroup := e.Group("/user-login-track")
	UseUserLoginTrackGroup(userLoginTrackGroup)

	websiteVisitGroup := e.Group("/website-visit")
	UseWebsiteVisitGroup(websiteVisitGroup)

	feedbackGroup := e.Group("/feedback")
	UseFeedbackRouter(feedbackGroup)

	userFavoriteGroup := e.Group("/user-favorite")
	UseUserFavoriteRouter(userFavoriteGroup)

	userSavedDishGroup := e.Group("/user-saved-dish")
	UseUserSavedDishRouter(userSavedDishGroup)

	userDietaryPreferenceGroup := e.Group("/user-dietary-preference")
	UseUserDietaryPreferenceRouter(userDietaryPreferenceGroup)

	userDishInteractionGroup := e.Group("/user-dish-interaction")
	UseUserDishInteractionRouter(userDishInteractionGroup)

	dishPopularityGroup := e.Group("/dish-popularity")
	UseDishPopularityRouter(dishPopularityGroup)

	userDishCollectionGroup := e.Group("/user-dish-collection")
	UseUserDishCollectionRouter(userDishCollectionGroup)

	personalizedRandomGroup := e.Group("/personalized-random")
	UsePersonalizedRandomRouter(personalizedRandomGroup)

	foodShopGroup := e.Group("/food-shop")
	UseFoodShopRouter(foodShopGroup)

	foodShopDishGroup := e.Group("/food-shop-dish")
	UseFoodShopDishRouter(foodShopDishGroup)

	userDishLocationGroup := e.Group("/user-dish-location")
	UseUserDishLocationRouter(userDishLocationGroup)

	notificationGroup := e.Group("/notification")
	UseNotificationRouter(notificationGroup)

}
