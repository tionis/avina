package uni_passau_bot

import (
	"encoding/csv"
	"github.com/jinzhu/now"
	"golang.org/x/text/encoding/charmap"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Global Variables
// Matrix Slice for food handling (should be replaced in future??)
var values [][]string

//var nextvalues [][]string

// UniPassauBot takes a telegram token and starts the uni passau bot on this bot account
func UniPassauBot(logger *slog.Logger, token string) {
	dbInit()

	botquit := make(chan bool) // channel for quitting of bot

	// catch os signals like sigterm and interrupt
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			logger.Info("Interruption Signal received, shutting down...")
			exit(logger, botquit)
		case syscall.SIGTERM:
			botquit <- true
		}
	}()

	// create bot object
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		logger.Error("Error starting Uni Passau Bot", err)
		return
	}

	// init reply keyboard
	replyBtn := tb.ReplyButton{Text: "Food for today"}
	replyBtn2 := tb.ReplyButton{Text: "Food for tomorrow"}
	replyBtn3 := tb.ReplyButton{Text: "Food for the week"}
	replyKeys := [][]tb.ReplyButton{
		{replyBtn, replyBtn2}, {replyBtn3}}

	// Command Handlers
	// handle special keyboard commands
	b.Handle(&replyBtn, func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			_, _ = b.Send(m.Sender, FoodToday(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
	})
	b.Handle(&replyBtn2, func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			_, _ = b.Send(m.Sender, FoodTomorrow(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
	})
	b.Handle(&replyBtn3, func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			_, _ = b.Send(m.Sender, FoodWeek(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
	})
	// handle standard text commands
	b.Handle("/hello", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Hi! How are you?", tb.ModeMarkdown)
		printInfo(logger, m)
	})
	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Hallo! Ich bin der inoffizielle ChatBot der Uni Passau! Was kann ich dir Gutes tun?\nWenn du Hilfe benötigst benutze einfach /help!\nSolltest du den Mensa- und Stundenplan in einer App wollen, schreibe /app für mehr Informationen", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
		printInfo(logger, m)
	})
	b.Handle("/app", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Du kannst dir die Android-App im [Play Store](https://play.google.com/store/apps/details?id=studip_uni_passau.femtopedia.de.unipassaustudip) gratis herunterladen.\nHinweis: Diese App wird von einer anderen Person entwickelt, bitte kontaktiere den App-Entwickler für Support!", tb.ModeMarkdown)
		printInfo(logger, m)

	})
	b.Handle("/help", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Information about the Bot is in the Description\nAvailable Commands are:\n*/help* - Show this help\n*/food* - Get Information for the food TODAY in the Uni Passau\n*/foodtomorrow* - Get Information for the food TOMORROW in the Uni Passau\n*/foodweek* - Get Information for the wood this WEEK in the Uni Passau\n*/contact* - Contact the bot maintainer for requests and bug reports\n*/app* - More Information for an useful Android-App for studip", tb.ModeMarkdown)
		printInfo(logger, m)
	})
	b.Handle("/food", func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, FoodToday(logger))
				logger.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, FoodToday(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
		//printAnswer(FoodToday(logger))
	})
	b.Handle("/foodtomorrow", func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, FoodTomorrow(logger))
				logger.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, FoodTomorrow(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
	})
	b.Handle("/foodweek", func(m *tb.Message) {
		if getTmp("uni-passau-bot", "isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, FoodWeek(logger))
				//_, _ = b.Send(m.Chat, "This command is temporarily disabled.")
				logger.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, FoodWeek(logger), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
				//_, _ = b.Send(m.Sender, "This command is temporarily disabled.")
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(logger, m)
	})
	b.Handle("/contact", func(m *tb.Message) {
		sendstring := ""
		if m.Text == "/contact" {
			_, _ = b.Send(m.Sender, "For requests and bug reports just add your message to the _/contact_ command.", tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Sender, "Sending Message to the Bot Maintainer...")
			tionis := tb.Chat{ID: 248533143}
			sendstring = "Message by " + m.Sender.FirstName + " " + m.Sender.LastName + "\nID: " + strconv.Itoa(int(m.Sender.ID)) + " Username: " + m.Sender.Username + "\n- - - - -\n" + strings.TrimPrefix(m.Text, "/contact ")
			_, _ = b.Send(&tionis, sendstring)
		}
		printInfo(logger, m)
		printAnswer(logger, sendstring)
	})
	b.Handle("/send", func(m *tb.Message) {
		if m.Sender.ID == 248533143 {
			s1 := strings.TrimPrefix(m.Text, "/send ")
			s := strings.Split(s1, "$")
			recID, _ := strconv.ParseInt(s[0], 10, 64)
			rec := tb.Chat{ID: recID}
			_, _ = b.Send(&rec, s[1])
		} else {
			_, _ = b.Send(m.Sender, "You are not authorized to execute this command!")
			printInfo(logger, m)
		}
	})
	b.Handle("Danke", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_Gern geschehen!_", tb.ModeMarkdown)
		printInfo(logger, m)
		printAnswer(logger, "_Gern geschehen!_")
	})
	b.Handle("Thanks", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_It's a pleasure!_", tb.ModeMarkdown)
		printInfo(logger, m)
		printAnswer(logger, "_It's a pleasure!_")
	})
	b.Handle("/ping", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_pong_", tb.ModeMarkdown)
		printInfo(logger, m)
		printAnswer(logger, "_pong_")
	})
	b.Handle(tb.OnAddedToGroup, func(m *tb.Message) {
		logger.Info("Group Message:")
		printInfo(logger, m)
	})
	b.Handle(tb.OnText, func(m *tb.Message) {
		sendstring := "Unknown Command - use help to get a list of available commands"
		if !m.Private() {
			logger.Info("Message from Group:")
		} else {
			_, _ = b.Send(m.Sender, sendstring)
		}
		printInfo(logger, m)
		printAnswer(logger, sendstring)
	})

	// Graceful Shutdown (botquit)
	go func() {
		<-botquit
		b.Stop()
		logger.Info("Bot was stopped")
		os.Exit(3)
	}()

	// init preparations
	loadFoodWeekArray(logger)

	// print startup message
	logger.Info("Starting up...")
	b.Start()
}

// FoodToday return a string of todays food
func FoodToday(logger *slog.Logger) string {
	// returns the string to print to user who requested the mensa plan
	// reads actual file
	err := updateFoodWeek(logger)
	if err != nil {
		return "An error occurred!"
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	t := time.Now().In(loc)
	ts := t.Format("02.01.2006")
	var daynum int

	for i := 1; i < 8; i++ {
		if weekDate(i) == ts {
			daynum = i
			break
		}
	}

	day := "*Essen am "

	switch daynum {
	case 1:
		day = day + "Montag:* 😋\n"
	case 2:
		day = day + "Dienstag:* 😋\n"
	case 3:
		day = day + "Mittwoch:* 😋\n"
	case 4:
		day = day + "Donnerstag:* 😋\n"
	case 5:
		day = day + "Freitag:* 😋\n"
	case 6, 7:
		day = "_Kein Essen heute!_ 😩"
	default:
		day = "An error occurred, please contact the administrator"
	}

	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(daynum) {
			if len(values[i]) >= 6 {
				day = day + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
			} else {
				day = day + "Error in this line\n"
			}
		}
	}

	return day
}

// FoodTomorrow returns a string for the food tomorrow
func FoodTomorrow(logger *slog.Logger) string {
	// returns the string to print to user who requested the mensa plan
	// reads actual file
	err := updateFoodWeek(logger)
	if err != nil {
		return "An error occurred!"
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	t := time.Now().In(loc)
	ts := t.Format("02.01.2006")
	var daynum int

	for i := 1; i < 8; i++ {
		if weekDate(i) == ts {
			daynum = i
			break
		}
	}
	daynum++
	if daynum == 8 {
		// Code not ready yet
		// TODO implement this with next week logic
		return "This only works weekdays, will be implemented soon!"
		/*
		   // Here Code for next week
		   loc, _ := time.LoadLocation("Europe/Berlin")
		   _, thisWeek := time.Now().In(loc).UTC().ISOWeek()
		   //nextweekstring := strconv.Itoa(thisWeek + 1)
		   //downloadFile(nextweekstring)
		   if thisWeek < 51 {
		       initNextArray()
		   } else {
		       return "Error! - Not implemented yet!"
		   }
		   // Verarbeitung
		   // Has to be monday as it would else trigger a another part of the code
		   daynum = 1
		   day := "*Essen am Montag:* 😋\n"
		   for i := 1; i < len(nextvalues); i++ {
		       if nextvalues[i][0] == nextWeekDate(daynum) {
		           if len(nextvalues[i]) >= 6 {
		               day = day + nextvalues[i][2] + ": " + delInf(nextvalues[i][3]) + " - " + transcor(nextvalues[i][6]) + " €\n"
		           } else {
		               day = day + "Error in this line\n"
		           }
		       }
		   }

		   return day*/
	} else if daynum > 8 {
		return "An Error occurred please contact the administrator"
	}

	day := "*Essen am "

	switch daynum {
	case 1:
		day = day + "Montag:* 😋\n"
	case 2:
		day = day + "Dienstag:* 😋\n"
	case 3:
		day = day + "Mittwoch:* 😋\n"
	case 4:
		day = day + "Donnerstag:* 😋\n"
	case 5:
		day = day + "Freitag:* 😋\n"
	case 6, 7:
		day = "_Kein Essen morgen!_ 😩"
	default:
		day = "An error occurred, please contact the administrator"
	}

	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(daynum) {
			if len(values[i]) >= 6 {
				day = day + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
			} else {
				day = day + "Error in this line\n"
			}
		}
	}

	return day
}

// FoodWeek returns a string of the food for the week
func FoodWeek(logger *slog.Logger) string {
	// reads actual file
	err := updateFoodWeek(logger)
	if err != nil {
		return "An error occurred!"
	}

	var Mo, Di, Mi, Do, Fr string
	Mo = "*Montag*:\n"
	Di = "*Dienstag*:\n"
	Mi = "*Mittwoch*:\n"
	Do = "*Donnerstag*:\n"
	Fr = "*Freitag*:\n"
	dayint := 1
	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(dayint) {
			switch dayint {
			case 1:
				if len(values[i]) >= 6 {
					Mo = Mo + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Mo = Mo + "Error in this line\n"
				}
			case 2:
				if len(values[i]) >= 6 {
					Di = Di + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Di = Di + "Error in this line\n"
				}
			case 3:
				if len(values[i]) >= 6 {
					Mi = Mi + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Mi = Mi + "Error in this line\n"
				}
			case 4:
				if len(values[i]) >= 6 {
					Do = Do + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Do = Do + "Error in this line\n"
				}
			case 5:
				if len(values[i]) >= 6 {
					Fr = Fr + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Fr = Fr + "Error in this line\n"
				}
			}
		} else {
			dayint++
		}
	}

	s := []string{Mo, Di, Mi, Do, Fr}
	return strings.Join(s, "\n")
}

// Time Calculation Logic

// WeekDate returns the date for an specific day
func weekDate(day int) string {
	// Start from the middle of the year:
	loc, _ := time.LoadLocation("Europe/Berlin")
	currentTime := time.Now().In(loc)
	_, week := time.Now().In(loc).UTC().ISOWeek()
	t := time.Date(currentTime.Year(), 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)
	ret := t.AddDate(0, 0, day-1)

	return ret.Format("02.01.2006")
}

/*func nextWeekDate(day int) string {
    // Start from the middle of the year:
    loc, _ := time.LoadLocation("Europe/Berlin")
    currentTime := time.Now().In(loc)
    _, week := time.Now().In(loc).UTC().ISOWeek()
    week++
    t := time.Date(currentTime.Year(), 7, 1, 0, 0, 0, 0, time.UTC)

    // Roll back to Monday:
    if wd := t.Weekday(); wd == time.Sunday {
        t = t.AddDate(0, 0, -6)
    } else {
        t = t.AddDate(0, 0, -int(wd)+1)
    }

    // Difference in weeks:
    _, w := t.ISOWeek()
    t = t.AddDate(0, 0, (week-w)*7)
    ret := t.AddDate(0, 0, day-1)

    return ret.Format("02.01.2006")
}*/

// Direct Data Manipulation Logic

// Load food for the week into array
func loadFoodWeekArray(logger *slog.Logger) {
	err := updateFoodWeek(logger)
	if err != nil {
		logger.Error("Could not update food for week: ", err)
	}
	loc, _ := time.LoadLocation("Europe/Berlin")
	_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
	weekstring := strconv.Itoa(thisWeek)
	r := csv.NewReader(strings.NewReader(getTmp("mensa", "food|week|"+weekstring)))
	values = nil
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		// Build Slice by appending every line
		values = append(values, record)
	}
}

// Load food for next week into array
/*func loadFoodNextWeekArray() {
    // Check if exists
    // checks if new file has to be downloaded and does so - does also remove the old file
    loc, _ := time.LoadLocation("Europe/Berlin")
    _, thisWeek := time.Now().In(loc).UTC().ISOWeek()
    nextWeekNumber := thisWeek + 1
    if nextWeekNumber > 52 {
        nextWeekNumber = 1
    }
    nextweekstring := strconv.Itoa(nextWeekNumber)
    if getTmp("mensa", "food|week"+nextweekstring) == "" {
        // No actual data found
        mensaBotLog.Info("No File for next week found - starting download")
        err := downloadFood(nextweekstring)
        if err != nil {
            mensaBotLog.Error("Could not download food for next week: ", err)
            return
        }
    }

    r := csv.NewReader(bufio.NewReader(strings.NewReader(getTmp("mensa", "food|week"+nextweekstring))))
    nextvalues = nil
    for {
        record, err := r.Read()
        // Stop at EOF.
        if err == io.EOF {
            break
        }
        // Build Slice by appending every line
        nextvalues = append(nextvalues, record)
    }
}*/

// As transformAndSaveFoodWeekData(input io.Reader) also changes all the commas in the prices to semicolons this func does the opposite
func transcor(input string) string {
	output := strings.ReplaceAll(input, ";", ",")
	return output
}

// Delete the symbols in the brackets at the end of the string (in this case the allergic info)
func delInf(input string) string {
	reg := regexp.MustCompile(`\(.*\)`)
	return reg.ReplaceAllString(input, "${1}")
}

// Transforms the data from the uni-passau version of the csv file to a standard one
func transformAndSaveFoodWeekData(logger *slog.Logger, input io.Reader, week string) error {
	// Transform data from ISO to UTF
	reader := charmap.ISO8859_1.NewDecoder().Reader(input)

	// Transforms csv file with separator ";" to a file with separator "," and also transforms all "," to ";"
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	if err != nil {
		logger.Error("Error reading from io.Reader to transform file: ", err)
		return err
	}
	newContents := strings.ReplaceAll(buf.String(), ",", "*")
	newContents = strings.ReplaceAll(newContents, ";", ",")
	newContents = strings.ReplaceAll(newContents, "*", ";")
	setTmp("mensa", "food|week|"+week, newContents, time.Until(now.EndOfWeek()))
	return nil
}

// DownloadFile downloads the newest file based on the week number
func downloadFood(logger *slog.Logger, week string) error {
	// Downloads the csv file
	s1 := []string{"https://www.stwno.de/infomax/daten-extern/csv/UNI-P/", week, ".csv"}
	url := strings.Join(s1, "")

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Could not download food for this week! Error: ", err)
		return err
	}

	// Load new data
	err = transformAndSaveFoodWeekData(logger, resp.Body, week)
	if err != nil {
		return err
	}
	loadFoodWeekArray(logger)
	err = resp.Body.Close()
	if err != nil {
		logger.Error("Error closing response Body: ", err)
		return err
	}
	return nil
}

// Update the food for the week
func updateFoodWeek(logger *slog.Logger) error {
	loc, _ := time.LoadLocation("Europe/Berlin")
	_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
	weekstring := strconv.Itoa(thisWeek)
	if getTmp("mensa", "food|week|"+weekstring) == "" {
		logger.Info("No File for this week found - starting download")
		err := downloadFood(logger, weekstring)
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop the program and kill hanging routines
func exit(logger *slog.Logger, quit chan bool) {
	// function for normal exit
	quit <- true
	simpleExit(logger)
}

// Exit while ignoring running routines
func simpleExit(logger *slog.Logger) {
	// Exit without using graceful shutdown channels
	logger.Info("Shutting down...")
	os.Exit(0)
}

// Print info regarding a given message
func printInfo(logger *slog.Logger, m *tb.Message) {
	logger.Info("[UniPassauBot] " + m.Sender.Username + " - " + m.Sender.FirstName + " " + m.Sender.LastName + " - ID: " + strconv.Itoa(int(m.Sender.ID)) + "Message: " + m.Text + "\n")
}

// Answer wrapper
func printAnswer(logger *slog.Logger, input string) {
	logger.Info("[UniPassauBot] Answer: " + input)
}

