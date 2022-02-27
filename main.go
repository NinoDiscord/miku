// 🌱 miku: Tiny, stateless microservice to notify that your Discord bot is going under maintenance, made in Go
// Copyright (c) 2022 Nino
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
)

var (
	// version returns the current version of Miku.
	version = "master"

	// commitSha returns the commit hash of when Miku was built.
	commitSha = "unknown"

	// buildDate returns the build date of when Miku was last built.
	buildDate = "???"
)

// MessageData represents the structure of using Go templates to customize
// the "undergoing maintenance" message that is sent once in all guilds.
type MessageData struct {
	// DiscordServer is the discord server to redirect users, this can be nil
	// if none was specified.
	DiscordServer string

	// Bot returns the bots username#discriminator
	Bot string
}

func init() {
	// Setup .env file is there is any
	if _, err := os.Stat("./.env"); err == nil || !os.IsNotExist(err) {
		if err := godotenv.Load("./.env"); err != nil {
			panic(err)
		}
	}

	// If the debug variable exists, let's put it into debug mode. :)
	if _, ok := os.LookupEnv("MIKU_DEBUG"); ok {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("Using v%s (commit=%s, built=%s) of Miku", version, commitSha, buildDate)
}

func main() {
	logrus.Info("Starting up miku...")

	message := "Bot **{{.Bot}}** is undergoing maintenance, please wait a bit!\nIf you need to know more, you can visit the Discord server: {{.DiscordServer}}!"
	var tmpl *template.Template
	var messageCache []string
	var currentUser string

	if m, ok := os.LookupEnv("MIKU_MESSAGE_TEMPLATE"); ok {
		t := template.New("miku template")
		t, err := t.Parse(m)
		if err != nil {
			logrus.Fatalf("Unable to parse Go template from `MIKU_MESSAGE_TEMPLATE` environment variable. %v", err)
		}

		message = m
		tmpl = t
	} else {
		t := template.New("miku default template")
		t, err := t.Parse(message)
		if err != nil {
			logrus.Fatalf("This should never happen; the default template was unable to be parsed: %v", err)
		}

		tmpl = t
	}

	// Check if we can grab the Discord token
	token, ok := os.LookupEnv("MIKU_DISCORD_TOKEN")
	if !ok {
		logrus.Fatalf("Missing `MIKU_DISCORD_TOKEN` environment variable.")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logrus.Fatal("Unable to create a Discord session:", err)
	}

	session.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		logrus.Debugf("Using Discord Gateway v%d", ready.Version)
		logrus.Infof("Successfully connected to Discord as %s.", fmt.Sprintf("%s#%s (%s)", ready.User.Username, ready.User.Discriminator, ready.User.ID))

		currentUser = fmt.Sprintf("%s#%s", ready.User.Username, ready.User.Discriminator)

		if err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: "dnd",
			Activities: []*discordgo.Activity{
				{
					Name: "the servers go whirrrr...",
					Type: 3,
				},
			},
		}); err != nil {
			logrus.Error("Unable to set presence:", err)
		}
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore bots
		if m.Author.Bot {
			return
		}

		// Skip if we do not have the `MIKU_PREFIX` environment variable,
		// so we do not spam guilds
		prefix, ok := os.LookupEnv("MIKU_PREFIX")
		if !ok {
			return
		}

		if !strContains(messageCache, m.GuildID) {
			if strings.HasPrefix(m.Content, prefix) {
				// Execute the Go template on the reader
				data := &MessageData{
					Bot:           currentUser,
					DiscordServer: os.Getenv("MIKU_DISCORD_SERVER"),
				}

				writer := bytes.NewBufferString("")
				if err := tmpl.Execute(writer, data); err != nil {
					logrus.Error("Unable to execute Go template:", err)
					return
				}

				if _, err := s.ChannelMessageSend(m.ChannelID, strings.Trim(writer.String(), " ")); err != nil {
					logrus.Error("Unable to send the message in channel:", err)
					return
				}

				messageCache = append(messageCache, m.GuildID)
			}
		}
	})

	// we only need guild messages
	session.Identify.Intents = discordgo.IntentsGuildMessages

	err = session.Open()
	if err != nil {
		logrus.Fatal("Unable to open a WebSocket connection:", err)
	}

	logrus.Info("We are now running! You should get logs if you're in debug mode! Press CTRL-C to exit~")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	logrus.Warn("We are now closing the connection...")
	if err := session.Close(); err != nil {
		logrus.Fatal("Unable to close session:", err)
	} else {
		logrus.Info("Goodbye... :(")
	}
}

func strContains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
