// Package vox manages the vox APIs
package vox

import (
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
	"disruptive/rest/vox/accounts"
	"disruptive/rest/vox/bank"
	"disruptive/rest/vox/characters"
	"disruptive/rest/vox/configs"
	"disruptive/rest/vox/demo"
	"disruptive/rest/vox/notifications"
	"disruptive/rest/vox/play"
	"disruptive/rest/vox/profiles"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	// Admin Accounts
	g := e.Group("/api/vox/accounts", auth.SetAdminMiddleware)
	g.GET("/:account_id", accounts.GetAccount)

	// Accounts
	g = e.Group("/api/vox/accounts", auth.SetMiddleware)
	g.POST("/emails", accounts.PostEmails)

	g.GET("/me", accounts.GetAccountMe)

	g.PATCH("/me", accounts.PatchAccountMe)
	g.DELETE("/me", accounts.DeleteAccountMe)

	g.GET("/me/bank/balance", accounts.GetBalance)
	g.GET("/me/bank/balance/available", accounts.GetAvailableBalance)
	g.PATCH("/me/bank/balance/add", accounts.AddBalance)
	g.PATCH("/me/bank/balance/charge", accounts.ChargeBank)
	g.POST("/me/bank/balance/gift_cards/:gift_card/redeem", accounts.RedeemGiftCard)

	g.GET("/me/bank/subscriptions", accounts.GetSubscription)
	g.PATCH("/me/bank/subscriptions/subscribe", accounts.Subscribe)
	g.PATCH("/me/bank/subscriptions/pending", accounts.Subscribe)
	g.DELETE("/me/bank/subscriptions/unsubscribe", accounts.Unsubscribe)

	g.POST("/me/email_pin", accounts.PostEmailPin)

	g.POST("/me/products/:product", accounts.PostAccountMeProduct)
	g.POST("/me/products/:product/:device_id/connect", accounts.PostAccountMeProductConnect)
	g.DELETE("/me/products/:product", accounts.DeleteAccountMeProduct)

	// to be removed
	g.POST("/me/characters/:product/:device_id/connect", accounts.PostAccountMeProductConnect)

	g.POST("/push-notifications", accounts.PostPushNotifications)

	g.GET("/preferences", accounts.GetPreferences)
	g.PUT("/preferences", accounts.PutPreferences)
	g.PATCH("/preferences", accounts.PatchPreferences)
	g.DELETE("/preferences", accounts.DeletePreferences)

	// Bank
	g = e.Group("/api/vox/bank")
	g.GET("/gift_cards", bank.GetGiftCards, auth.SetMiddleware)
	g.GET("/gift_cards/:gift_card", bank.GetGiftCard, auth.SetMiddleware)
	g.DELETE("/gift_cards/:gift_card", bank.ExpireGiftCard, auth.SetMiddleware)

	g.POST("/android/iap/transaction", bank.PostAndroidIAPTransaction, auth.SetMiddleware)
	g.POST("/android/sub/transaction", bank.PostAndroidSubTransaction, auth.SetMiddleware)
	g.POST("/android/notifications", bank.PostAndroidNotifications)

	g.POST("/apple/iap/transaction", bank.PostAppleIAPTransaction, auth.SetMiddleware)

	// Characters
	g = e.Group("/api/vox/characters", auth.SetMiddleware)
	g.GET("", characters.GetCharacters)

	// Configuration
	g = e.Group("/api/vox/configs", auth.SetMiddleware)
	g.GET("/bank/rates", bank.GetRates)
	g.GET("/bank/skus", bank.GetSKUs)
	g.GET("/characters", characters.GetCharacters)
	g.GET("/loadbalancers", configs.GetLoadBalancers)
	g.GET("/loadbalancers/:country/:name", configs.GetLoadBalancer)
	g.GET("/localize/:version/:language", configs.GetLocalize)
	g.GET("/products", configs.GetProducts)
	g.GET("/products/:product", configs.GetProduct)
	g.GET("/tti/styles", play.GetTTIStyles)

	// Moderation
	g = e.Group("/api/vox/moderate", auth.SetMiddleware)
	g.POST("", play.PostModerate)

	// Profiles
	g = e.Group("/api/vox/profiles", auth.SetMiddleware)
	g.GET("", profiles.GetAll)
	g.POST("", profiles.Post)

	g.GET("/:profile_id", profiles.Get)
	g.PATCH("/:profile_id", profiles.Patch)
	g.DELETE("/:profile_id", profiles.Delete)

	g.PATCH("/:profile_id/characters/:character_version", profiles.PatchCharacter)

	g.GET("/:profile_id/characters/:character_version/archives", profiles.GetArchiveIndex)
	g.GET("/:profile_id/characters/:character_version/archives/:archive_id", profiles.GetArchiveByID)
	g.GET("/:profile_id/characters/:character_version/archives/entries/date_range", profiles.GetArchiveEntriesByDateRange)
	g.GET("/:profile_id/characters/:character_version/archives/summaries/date_range", profiles.GetArchiveSummaryByDateRange)
	g.GET("/:profile_id/characters/:character_version/archives/entries/:session_id", profiles.GetSessionEntryByID)
	g.DELETE("/:profile_id/characters/:character_version/archives/summaries/date_range", profiles.DeleteArchiveSummaryDateRange)
	g.DELETE("/:profile_id/characters/:character_version/memory", profiles.DeleteSessionMemory)

	g.GET("/:profile_id/picture", profiles.GetProfilePicture)
	g.PUT("/:profile_id/picture", profiles.PutProfilePicture)

	g.GET("/:profile_id/preferences", profiles.GetPreferences)
	g.PUT("/:profile_id/preferences", profiles.PutPreferences)
	g.PATCH("/:profile_id/preferences", profiles.PatchPreferences)
	g.DELETE("/:profile_id/preferences", profiles.DeletePreferences)

	// Play Character
	g = e.Group("/api/vox/play/:profile_id/:character_version", auth.SetMiddleware)
	g.POST("/sts/audio/file", play.PostSTSAudioFile)
	g.POST("/sts/audio/stream", play.PostSTSAudioStream)
	g.POST("/sts/text", play.PostSTSText)
	g.POST("/sts/text/predefined", play.PostSTSTextPredefined)
	g.GET("/sts/audio/:audio_id", play.GetSTSAudio)
	g.GET("/sts/audio/ws/:audio_id", play.GetSTSAudioWS)
	g.GET("/sts/moderation/:audio_id", play.GetSTSModeration)
	g.GET("/sts/text/:audio_id", play.GetSTSText)

	g.POST("/sts/close/:audio_id", play.PostSTSClose)
	g.PATCH("/sts/end_sequence", play.PatchSTSEndSequence)

	g.GET("/ttt", play.GetTTT)

	g.GET("/audio/:session_id/assistant", play.GetAssistantAudio)
	g.GET("/audio/:session_id/user", play.GetUserAudio)

	g.GET("/tti", play.GetTTI)

	// Admin Notifications
	g = e.Group("/api/vox/notifications", auth.SetAdminMiddleware)
	g.DELETE("/:notification_id", notifications.Delete)

	// Notifications
	g = e.Group("/api/vox/notifications", auth.SetMiddleware)
	g.GET("", notifications.GetAll)
	g.POST("", notifications.Post)
	g.GET("/:notification_id", notifications.Get)
	g.PATCH("/:notification_id", notifications.Patch)

	// Demo
	g = e.Group("/api/vox/demo")
	g.PATCH("/:user", demo.PatchUser)
}
