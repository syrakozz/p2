package notifications

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"sort"
	"strings"
	"time"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"
)

const notificationHTML = `
<!DOCTYPE html>
<html>
<head>
<style>    
  blockquote {
    margin-left: 20px;
    border-left: 2px solid #333;
    padding-left: 10px;
  }
</style>
</head>
<body>
  <h3>{{.Title}}</h3>
  {{if .AssessmentTranslation}} 
  <h4>{{.AssessmentLabel}}:</h4>
  <li>{{.AssessmentTranslation}}</li>
  {{end}}
  <h4>{{.Triggered}}:</h4>
  <ul>
  	{{range $v := .Categories}}
	<li>{{$v}}</li>
	{{end}}
	{{if .Toxic}}
    <li>{{.ToxicLabel}}</li>
    {{end}}
	{{if .NotAgeAppropriate}}
    <li>{{.NotAgeAppropriateLabel}}</li>
    {{end}}
  </ul>
  <p>

  <h4>{{.InformationSection}}</h4>
  <ul>
    <li>{{.ProfileLabel}}: {{.Name}}</li>
    <li>{{.CharacterLabel}}: {{.Character}}</li>
    <li>{{.TimeLabel}}: {{.Time}}</li>
  </ul>

  <h4>{{.ContentSection}}</h4>

  <blockquote>
    <h4>{{.Name}}</h4>
    {{.User}}
  </blockquote>
<body>
</html>
`

type notificationHTMLVars struct {
	AssessmentLabel        string
	AssessmentTranslation  string
	AllowCharacterResponse bool
	Assistant              string
	Categories             []string
	Character              string
	CharacterLabel         string
	ContentSection         string
	InformationSection     string
	Name                   string
	NotAgeAppropriate      bool
	NotAgeAppropriateLabel string
	ProfileLabel           string
	RepliedSection         string
	Time                   string
	TimeLabel              string
	Title                  string
	Toxic                  bool
	ToxicLabel             string
	Triggered              string
	User                   string
}

// SendModerationEmail sends a moderation email.
func SendModerationEmail(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterName string, lastUserAudio *characters.UserAudio, sessionID int, localize *configs.Localize) error {
	fid := slog.String("fid", "vox.notifications.SendModerationEmail")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	t, err := template.New("t").Parse(notificationHTML)
	if err != nil {
		logCtx.Error("unable to parse template", fid, "error", err)
		return err
	}

	var buf bytes.Buffer

	loc, err := time.LoadLocation(account.Timezone)
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}

	vars := notificationHTMLVars{
		AssessmentTranslation:  lastUserAudio.Moderation.Analysis.AssessmentTranslation,
		AssessmentLabel:        localize.Email["moderation_notification_assessment_label"],
		NotAgeAppropriate:      lastUserAudio.Moderation.Analysis.NotAgeAppropriate,
		NotAgeAppropriateLabel: localize.TextAnalysis["not_age_appropriate"],
		Character:              characterName,
		CharacterLabel:         localize.Email["moderation_notification_character_label"],
		ContentSection:         localize.Email["moderation_notification_content_section"],
		InformationSection:     localize.Email["moderation_notification_information_section"],
		Name:                   profile.Name,
		ProfileLabel:           localize.Email["moderation_notification_profile_label"],
		Time:                   lastUserAudio.Timestamp.In(loc).Format(time.DateTime),
		TimeLabel:              localize.Email["moderation_notification_time_label"],
		Title:                  fmt.Sprintf(localize.Email["moderation_notification_title"], profile.Name, characterName),
		Toxic:                  lastUserAudio.Moderation.Analysis.Toxic,
		ToxicLabel:             localize.TextAnalysis["toxic"],
		Triggered:              localize.Email["moderation_notification_triggered"],
		User:                   lastUserAudio.Text,
	}

	for k, v := range lastUserAudio.Moderation.Categories {
		if !v {
			continue
		}

		vars.Categories = append(vars.Categories, localize.Moderation[k])
	}

	sort.Strings(vars.Categories)
	if err := t.Execute(&buf, vars); err != nil {
		logCtx.Error("unable to execute email template", fid, "error", err)
		return err
	}

	subject := fmt.Sprintf(localize.Email["moderation_notification_subject"], sessionID, characterName, profile.Name)
	if sessionID == 0 {
		strings.Replace(subject, " #0 ", " ", 1)
	}

	req := accounts.EmailRequest{
		From:    config.VARS.MailgunNotificationsFrom,
		To:      profile.Notifications.Emails,
		Subject: subject,
		HTML:    buf.String(),
	}

	return accounts.SendEmail(ctx, logCtx, req)
}

const pinHTML = `
<!DOCTYPE html>
<html>
<body>
  <h2>{{.Title}}</h2>
  {{.YourPinCode}}: <b>{{.Pin}}</b>
  {{.Body}}
<body>
</html>
`

type pinHTMLVars struct {
	Body        template.HTML
	Pin         string
	Title       string
	YourPinCode string
}

// SendPinEmail sends the current account pin to the account email address.
func SendPinEmail(ctx context.Context, logCtx *slog.Logger, language string) error {
	fid := slog.String("fid", "vox.notifications.SendPinEmail")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	if account.Email == "" {
		return nil
	}

	localize, err := configs.GetLocalization(ctx, logCtx, "v1", language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return err
	}

	t, err := template.New("t").Parse(pinHTML)
	if err != nil {
		logCtx.Error("unable to parse template", fid, "error", err)
		return err
	}

	pin := localize.Email["pin_notification_pin_not_set"]
	if account.Pin != "" {
		pin = account.Pin
	}

	var buf bytes.Buffer

	vars := pinHTMLVars{
		Body:        template.HTML(localize.Email["pin_notification_body"]),
		Pin:         pin,
		Title:       localize.Email["pin_notification_title"],
		YourPinCode: localize.Email["pin_notification_your_pin_code"],
	}

	if err := t.Execute(&buf, vars); err != nil {
		logCtx.Error("unable to execute email template", fid, "error", err)
		return err
	}

	req := accounts.EmailRequest{
		From:    config.VARS.MailgunNotificationsFrom,
		To:      []string{account.Email},
		Subject: localize.Email["pin_notification_subject"],
		HTML:    buf.String(),
	}

	if err := accounts.SendEmail(ctx, logCtx, req); err != nil {
		logCtx.Error("unable to send mail", fid, "error", err)
		return err
	}

	return nil
}
