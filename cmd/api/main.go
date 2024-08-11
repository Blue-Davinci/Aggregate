package main

import (
	"database/sql"
	"expvar"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/jsonlog"
	"github.com/blue-davinci/aggregate/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

// a quick variable to hold our version. ToDo: Change this.
const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	scraper struct {
		noofroutines  int
		fetchinterval int
		scraperclient struct {
			retrymax int
			timeout  int
		}
	}
	notifier struct {
		cronJob        *cron.Cron
		interval       int64
		deleteinterval int64
	}
	cors struct {
		trustedOrigins []string
	}
	paystack struct {
		cronJob                                    *cron.Cron
		secretkey                                  string
		initializationurl                          string
		verificationurl                            string
		chargeauthorizationurl                     string
		autosubscriptioninterval                   int64
		checkexpiredsubscriptioninterval           int64
		checkexpiredchallengedsubscriptioninterval int64
	}
	frontend struct {
		baseurl          string
		activationurl    string
		passwordreseturl string
		callback_url     string
	}
	limitations struct {
		maxFeedsCreated  int
		maxFeedsFollowed int
		maxComments      int
	}
}
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	//Initialize our get ENV for our DSN
	getCurrentPath(logger)
	// Initialize the flags
	var cfg config
	//initFlags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Database Flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("AGGREGATE_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	// Our SMTP flags with given defaults.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "53aa513750477d", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "15eb41b4f34521", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Aggregate <no-reply@aggregate.com>", "SMTP sender")
	// Scraper settings
	flag.IntVar(&cfg.scraper.noofroutines, "scraper-routines", 5, "Number of scraper routines to run")
	flag.IntVar(&cfg.scraper.fetchinterval, "scraper-interval", 40, "Interval in seconds before the next bunch of feeds are fetched")
	flag.IntVar(&cfg.scraper.scraperclient.retrymax, "scraper-retry-max", 3, "Maximum number of retries for HTTP requests")
	flag.IntVar(&cfg.scraper.scraperclient.timeout, "scraper-timeout", 15, "HTTP client timeout in seconds")
	// Payment
	flag.StringVar(&cfg.paystack.secretkey, "paystack-secret", os.Getenv("PAYSTACK_SECRET_KEY"), "Paystack Secret Key")
	flag.StringVar(&cfg.paystack.initializationurl, "paystack-initialization-url", "https://api.paystack.co/transaction/initialize", "Paystack Initialization URL")
	flag.StringVar(&cfg.paystack.verificationurl, "paystack-verification-url", "https://api.paystack.co/transaction/verify/", "Paystack Verification URL")
	flag.StringVar(&cfg.paystack.chargeauthorizationurl, "paystack-charge-authorization-url", "https://api.paystack.co/transaction/charge_authorization", "Paystack Charge Authorization URL")
	flag.Int64Var(&cfg.paystack.autosubscriptioninterval, "paystack-autosubscription-interval",
		720, "Interval in minutes for the auto subscription") // run auto-subscription checks every 12 hours
	flag.Int64Var(&cfg.paystack.checkexpiredsubscriptioninterval, "paystack-check-expired-subscription-interval",
		1440, "Interval in minutes for the check expired subscription") // run check expired subscription every 24 hours
	flag.Int64Var(&cfg.paystack.checkexpiredchallengedsubscriptioninterval, "paystack-check-expired-challenged-subscription-interval",
		720, "Interval in minutes for the check expired challenged subscription") // run check expired challenged subscription every 12 hours
	// Read the frontend url into the config struct
	flag.StringVar(&cfg.frontend.baseurl, "frontend-url", "http://localhost:5173", "Frontend URL")
	flag.StringVar(&cfg.frontend.activationurl, "frontend-activation-url", "http://localhost:5173/verify?token=", "Frontend Activation URL")
	flag.StringVar(&cfg.frontend.passwordreseturl, "frontend-password-reset-url", "http://localhost:5173/reset/password?token=", "Frontend Password Reset URL")
	flag.StringVar(&cfg.frontend.callback_url, "frontend-callback-url", "https://adapted-healthy-monitor.ngrok-free.app/v1", "Frontend Callback URL")
	// Limitations
	flag.IntVar(&cfg.limitations.maxFeedsCreated, "max-feeds-created", 5, "Maximum number of feeds a non-registered user can create")
	flag.IntVar(&cfg.limitations.maxFeedsFollowed, "max-feeds-followed", 5, "Maximum number of feeds a non-registered user can follow")
	flag.IntVar(&cfg.limitations.maxComments, "max-comments", 10, "Maximum number of comments a non-registered user can make")
	// Cors
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	// fetching interval for the notifier
	flag.Int64Var(&cfg.notifier.interval, "notifier-interval", 10, "Interval in minutes for the notifier to fetch new notifications")
	// delete interval for the notifier
	flag.Int64Var(&cfg.notifier.deleteinterval, "notifier-delete-interval", 100, "Interval in minutes for the notifier to delete old notifications")
	//parse our flags
	flag.Parse()
	// initialize our cron jobs
	cfg.paystack.cronJob = cron.New()
	cfg.notifier.cronJob = cron.New()
	// create our connection pull
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("database connection pool established", nil)
	// Init our exp metrics variables for server metrics.
	publishMetrics()
	// setup our application with all dependancies Injected.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// start our background workers
	app.startBackgroundWorkers()
	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func (app *application) startBackgroundWorkers() {
	// start the scraper function for our RSSFeeds as a goroutine
	go app.startRssFeedScraperHandler()
	// hook our notifier
	go app.fetchNotificationsHandler()
	// start our server
	go app.startPaymentSubscriptionHandler()
}

// publishMetrics sets up the expvar variables for the application
// It sets the version, the number of active goroutines, and the current Unix timestamp.
func publishMetrics() {
	expvar.NewString("version").Set(version)
	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}

// getCurrentPath invokes getEnvPath to get the path to the .env file based on the current working directory.
// After that it loads the .env file using godotenv.Load to be used by the initFlags() function
func getCurrentPath(logger *jsonlog.Logger) string {
	currentpath := getEnvPath(logger)
	if currentpath != "" {
		err := godotenv.Load(currentpath)
		if err != nil {
			logger.PrintFatal(err, nil)
		}
	} else {
		logger.PrintFatal(nil, map[string]string{
			"error": "unable to load .env file",
		})
	}
	logger.PrintInfo("Loading Environment Variables", map[string]string{
		"DSN": currentpath,
	})
	return currentpath
}

// openDB() opens a new database connection using the provided configuration.
// It returns a pointer to the sql.DB connection pool and an error value.
func openDB(cfg config) (*database.Queries, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// Use ping to establish new conncetions
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	queries := database.New(db)
	return queries, nil
}
