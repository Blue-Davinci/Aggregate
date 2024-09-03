<p align="center">
  <a href="" rel="noopener">
 <img width=200px height=200px src="https://i.ibb.co/WKxXnqw/agglogo.png" alt="Aggregate logo"></a>
</p>

<h3 align="center">Aggregate</h3>

<div align="center">

[![Status](https://img.shields.io/badge/status-active-success.svg)]()
[![GitHub Issues](https://img.shields.io/github/issues/Blue-Davinci/Aggregate.svg)](https://github.com/Blue-Davinci/Aggregate/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/Blue-Davinci/Aggregate.svg)](https://github.com/Blue-Davinci/Aggregate/pulls)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)

</div>

<hr />

<p align="center"> Aggregate [Back-End]: This is the full Golang backend code for the <strong>AGGREGATE Project</strong> 🚀🌐

<hr />

You are currently in the BackEnd section, To view the FrontEnd go [here](https://github.com/Blue-Davinci/Aggregate-FronteEnd)
    <br> 
</p>

## 📝 Table of Contents

- [About](#about)
- [Features](#features)
- [Getting Started](#getting_started)
- [Deployment](#deployment)
- [Usage](#usage)
- [Algo](#algo)
- [Payment](#payment)
- [Built Using](#built_using)
- [API Endpoints](#endpoints)
- [TODO](./TODO.md)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

<hr />

## 🧐 About <a name = "about"></a>

Aggregate is a content aggregation platform designed to streamline information consumption. Its purpose is to centralize feeds from various sources—such as RSS and Atom—into a unified stream. Users can effortlessly follow their favorite content, whether it’s news, blogs, or other updates. The project emphasizes efficiency, security, and a user-friendly experience, making it a valuable tool for staying informed in today’s fast-paced digital landscape. 🚀🌐

## ✨ Features <a name="features"></a>

This will be a high-level list of the features this API supports. For a more detailed and in-depth list of features, you can visit the frontend [here](https://github.com/Blue-Davinci/Aggregate-FronteEnd?tab=readme-ov-file#-usage-and-features-).

Some of the features include:

1. **Administration\Admin endpoints:**
   - Provides Endpoints to facilitate admin activities and management. Allows only users with Admin permissions to perform operations such as:
        - User management: view all users and their activities
        - Subscription Management: add\hide\update plans, subscriptions.
        - Payment Plan Management: Manage your available plans.
        - Permission Management: add\remove\update and get all available permissions.
        - Statistics : General endpoint showing general API statistics.
        - Errors: Allows admins to monitor the scraper section of the API.
        - Announcements: Allows admins to manage announcements.
    **More capabilities are in the pipeline including feed and post management as well as Moderation**
2. **Permissions:**
   - Admins can set and manage types of permissions as well as individual permissions.
   - With the above capability, permissions become highly customizeable in that  you can further specify which routes require which permissions for example,
     you may have `{comment:write}` and `{comment:read}`, if a moderator bans a user, their `commen:write` permission maybe removed, and thus the users
     replies and comments will not be reflected.

3. **Scraper:**
    - A custom RSS scraper designed to scrape all supported rss feed types including Atom feeds
    - Uses custom retryable client, as well as flags to allow customizations including timeouts and retries.

4. **Sanitization:**
    - Sanitizes all relevanct rss fields from the scraped links. It offers both strict sanitization as well as a 'gentle' one that doesn't strip every HTML.

5. **Panic, Shutdown, and Recovery**: 
   - The API supports shutdown and panic recoveries, including wait times and graceful shutdown procedures which support background routines and cron jobs.

6. **CORS Management**: 
   - Support for CORS management, including setting authorized/permitted URLs, methods, and more.

7. **Metrics**: 
   - The API supports metrics, allowing authorized users to view items such as the number of goroutines, connection pools, performance parameters, and many others.

8. **Mailer Support**: 
   - The API supports email sending and template development. All you need to do is hook up your `smtp` settings as shown in the `flags` section.

9. **Rate Limiter**: 
   - The API allows you to set limitations on the rates of user requests, with flags you can set using the listed entries in the `flags` section.

10. **Custom Login and Authentication**: 
   - The API uses a custom authentication integration, not OAuth or JWT, but rather a hybrid of bearer tokens and an API key for better security.

11. **Filtering, Sorting, and Pagination Support**: 
   - Most of the routes and handlers allow the usage of filters as well as pagination.

12. **Custom JSON Logger**: 
   - The API uses a custom structured JSON logger with support for stack traces and customizations depending on the type of info being outputted.

13. **Fully Functional Scraper**: 
   - The API uses its own scraper designed to acquire all flagged aggregated feeds and has multiple flags such as client timeouts, feed types, scraping rates, retry rates, etc., all designed to be customizable as well as fast.

14. **Payment Client Support**: 
    - The API provides custom clients for integration with the payment gateway.

15. **Customization & Flexibility**:
    - The API provides a myriad of flags that a user can use to adjust and manipulate the API to their own needs. From setting your own rate limits, to how frequently the cron jobs run and happen, to the timeouts to be considered during a scrape job, to how many units to scrape per x time, the API gives you all the power you need to make it yours.

16. **Front-End Support and Flexibility**: 
    - The API provides provision for users to integrate various links and **EMAIL** templates to their own frontend locations. Whether it's password reset links to Payment redirections to Activation links etc, users can customize and integrate their frontend endpoints with this API.

**In The Works:**
- For more of what is currently being worked on, please Visit the [Todo](./TODO.md)

**NB: For a full list of all the capabilities and expressed better in a frontend, please VISIT the aggregate frontend, linked above this page.**

## 🏁 Getting Started <a name = "getting_started"></a>

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See [deployment](#deployment) for notes on how to deploy the project on a live system.

### Prerequisites

What things you need to install the software and how to install them. See [deployment](#deployment) for notes on how to deploy the project on a live system.

### Prerequisites

Before you can run or contribute to this project, you'll need to have the following software installed:

- [Go](https://golang.org/dl/): The project is written in Go, so you'll need to have Go installed to run or modify the code.
- [PostgreSQL](https://www.postgresql.org/download/): The project uses a PostgreSQL database, so you'll need to have PostgreSQL installed and know how to create a database.
- A Go IDE or text editor: While not strictly necessary, a Go IDE or a text editor with Go support can make it easier to work with the code. I use vscode.
- [Git](https://git-scm.com/downloads): You'll need Git to clone the repo.

<hr />

### Installing

1. **Clone the repository:** Start by cloning the repository to your local machine. Open a terminal, navigate to the directory where you want to clone the repository, and run the following command:

    ```bash
    git clone https://github.com/blue-davinci/aggregate.git
    ```

2. **Navigate to the project directory:** Use the `cd` command to navigate to the project directory:

    ```bash
    cd aggregate
    ```

3. **Install the Go dependencies:** The Go tools will automatically download and install the dependencies listed in the `go.mod` file when you build or run the project. To download the dependencies without building or running the project, you can use the `go mod download` command:

    ```bash
    go mod download
    ```
4. **Set up the database:** The project uses a PostgreSQL database. You'll need to create a new database and update the connection string in your configuration file or environment variables.
We use `GOOSE` for all the data migrations and `SQLC` as the abstraction layer for the DB. To proceed
with the migration, navigate to the `Schema` director:

```bash
cd internal\sql\schema
```
- Then proceed by using the `goose {connection string} up` to execute an <b>Up migration</b> as shown:
- <b>Note:</b> You can use your own environment variable or load it from the env file.

```bash
goose postgres postgres://aggregate:password@localhost/aggregate  up
```

5. **Build the project:** You can build the project using the makefile's command:

    ```bash
    make build/api
    ```
    This will create an executable file in the current directory.
    <b>Note: The generated executable is for the windows environment</b>
      <b>- However, You can find the linux build command within the makefile!</b>

6. **Run the project:** You can run the project using the `go run` or use <b>`MakeFile`</b> and do:

    ```bash
    make run/api
    ```

7. **MakeFile Help:** For additional supported commands run `make help`:

  ```makefile
  make help
  ```

  **Output:**
  ```
  make help
  Usage: 
  run/api            -  run the api application
  db/psql            -  connect to the db using psql
  build/api          -  build the cmd/api application
  audit              -  tidy dependencies and format, vet and test all code
  db/migrations/up   -  run the up migrations using confirm as prerequisite
  vendor             -  tidy and vendor dependencies
  ```

### Description

The application accepts command-line flags for configuration, establishes a connection pool to a database, and publishes variables for monitoring the application. The published variables include the application version, the number of active goroutines and the current Unix timestamp.
  - This will start the application. You should be able to access it at `http://localhost:4000`.

<hr />

## Optional Parameters <a name = "optpars"></a>

You can view the **parameters** by utilizing the `-help` command. Here is a rundown of 
the available commands for a quick lookup.
- **smtp-sender:** Sets the sender for SMTP (email) communications. Default: "Aggregate <no-reply@aggregate.com>".
- **cors-trusted-origins [value]:** Trusted CORS origins (space separated)
- **db-dsn [string]:** PostgreSQL DSN (default "{Path to your .env holding your DSN}")
- **db-max-idle-conns [int]:** PostgreSQL max idle connections (default 25)
- **db-max-idle-time [string]:** PostgreSQL max connection idle time (default "15m")
- **db-max-open-conns [int]:** PostgreSQL max open connections (default 25)
- **env [string]:** Environment (development|staging|production) (default "development")
- **port [int]:** API server port (default 4000)
- **smtp-host [string]:** SMTP host (default "sandbox.smtp.mailtrap.io"- I use mailtrap for tests)
- **smtp-password [string]:** SMTP password (default "xxxxx")
- **smtp-port [int]:** SMTP port (default 25)
- **smtp-sender [string]:** SMTP sender (default "Aggregate <no-reply@aggregate.com>")
- **smtp-username [string]:** SMTP username (default "skunkhunt42")
- **baseurl [string]:** frontend url (default "http://localhost:5173")
- **activationurl [string]:** frontend activation url (default "http://localhost:5173/verify?token=")
- **passwordreseturl:** frontend password reset url (default "http://localhost:5173/reset?token=")
- **scraper-routines [int]:** Number of scraper routines to run (default 5)- **scraper-interval [int]:** Interval in seconds before the next bunch of feeds are fetched (default 40)
- **scraper-retry-max [int]:** Maximum number of retries for HTTP requests (default 3)
- **scraper-timeout [int]:** HTTP client timeout in seconds (default 15)
- **~~cors-trusted-origins [value]~~:** ~~Trusted CORS origins (space separated)~~
- **notifier-interval [int64]:** Interval in minutes for the notifier to fetch new notifications (default 10)
- **notifier-delete-interval [int64]:** Interval in minutes for the notifier to delete old notifications (default 100)
- **callback_url [string]:** Represents the url which the payment gateway will navigate to after a transaction.
- **maxFeedsCreated [int64]:** A limitation flag that sets the max number of feeds a free tier user can create
- **maxFeedsFollowed [int64]:** A limitation flag that sets the max number of feeds a free tier user can follow
- **maxComments [int64]:** A limitation flag that sets the max number of comments a free tier user can make
- **limiter-burst [int]:** Rate limiter maximum burst (default 4)
- **limiter-enabled [bool]:** Enable rate limiter (default true)
- **limiter-rps [float]:** Rate limiter maximum requests per second (default 2)
- **paystack-autosubscription-interval [int]:** Interval in minutes for the auto subscription (default 720)
- **paystack-charge-authorization-url [string]:** The Paystack Charge Authorization URL for processing recurring charges.
- **paystack-check-expired-challenged-subscription-interval [int]:** Interval in minutes for the check on expired challenged subscription 
- **paystack-check-expired-subscription-interval [int]:** Interval in minutes for the check on expired subscription.
- **paystack-initialization-url [string]:** The Paystack Initialization URL for processing the initialization of a payment transaction
- **paystack-secret [string]:** Paystack Secret Key. This can be configured above, see [payment configuration here](#payment)
- **paystack-verification-url [string]:** Paystack Verification URL endpoint to process the payment verifications.
- **sanitization-strict [bool]:** allows a user to specify the level of sanitization. Setting this as true will be equivalent to stripping all `HTML` and all their `attributes`. The default is false for a medium balance.

Using `make run`, will run the API with a default connection string located 
in `cmd\api\.env`. If you're using `powershell`, you need to load the values otherwise you will get
a `cannot load env file` error. Use the PS code below to load it or change the env variable:

```powershell
$env:AGGREGATE_DB_DSN=(Get-Content -Path .\cmd\api\.env | Where-Object { $_ -match "AGGREGATE_DB_DSN" } | ForEach-Object { $($_.Split("=", 2)[1]) })
```

Alternatively, in unix systems you can make a .envrc file and load it directly in the makefile by importing like so:
```makefile
include .envrc
```

A succesful run will output something like this:

```bash
make run/api
'Running cmd/api...'
go run ./cmd/api
{"level":"INFO","time":"2024-08-26T16:10:34Z","message":"Loading Environment Variables","properties":{"DSN":"cmd\\api\\.env"}}
{"level":"INFO","time":"2024-08-26T16:10:34Z","message":"database connection pool established"}
{"level":"INFO","time":"2024-00-26T16:00:34Z","message":"Starting RSS Feed Scraper","properties":{"Client Timeout":"15","Interval":"40s","No of Go Routines":"5","No of Retries":"3"}}
{"level":"INFO","time":"2024-00-26T16:00:34Z","message":"Starting autosubscription jobs...","properties":{"Interval":"720"}}
```

<hr />

### API Endpoints <a name = "endpoints"></a>
Below are most of the accepted flags in the API and a high level description of what they do. To view the comprehensive list please run the application with the `-help` flag:

1. **GET /v1/healthcheck:** Checks the health of the application. Returns a 200 OK status code if the application is running correctly.

2. **POST /v1/users:** Registers a new user.

3. **PUT /v1/users/activated:** Activates a user.

4. **POST /v1/api/authentication:** Creates an authentication token.

5. **GET /debug/vars:** Provides debug variables from the `expvar` package. 

6. **POST /feeds:** Add an RSS Type feed {Atom/RSS}

7. **GET /feeds?page=1&page_size=30:** Get all Feeds in the DB, <b>With Pagination</b>
    <b>Note:</b> <i>You can leave the pagination parameters foe default values!</i>

8. **POST /feeds/follows:** Follow any feed for a user

9. **GET /feeds/follow:** Get all feeds followed by a user. <b>Supports pagination and search</b>.

10. **DELETE /feeds/follow/{feed_id}:** Unfollow a followed feed

11. **GET /feeds:** Get all Posts from scraped feeds that are followed by a user. Supports pagination and search.

12. **POST /password-reset:** Initial request for password reset that sends a validation tokken

13. **PUT /password:** Updates actual password after reset.

14. **GET /notifications:** Retrieve notifications on per user basis. Current implimentation supports <b>polling and on-demand basis</b>

15. **GET /feeds/favorites:** Get favorite feeds for a user. <b>Supports pagination and search</b>.

16. **POST /feeds/favorites:** Add a new favorite post

17. **DELETE /feeds/favorites:** Deletes/Remove a favorited post

18. **GET /feeds/favorites/post:** Gets detailed post infor for only favorited posts i.e Can see any favorited content. <b>Supports pagination and search</b>.

19.  **GET /feeds/follow/list:** Gets the list of all feeds followed by a user.<b>Supports pagination and search</b>.

20. **GET /feeds/follow/posts/comments:** Gets all comments related to a specific post

21. **POST /feeds/follow/posts/comments:** Post a comment from a user

22. **PATCH /feeds/created/{}:** Feed Manager. Allows a user to edit a feed they created. Allows hiding and unhiding of created feeds.

23. **GET /feeds/created:** Feed Manager. Get all feeds created by a user as well as related statistics such as follows and ratings.

24. **GET /follow/posts/comments/{postID}:** Get all comments for a particular post

25. **DELETE /follow/posts/comments/{postID}:** Remove/clear a comment notification

26. **POST /follow/posts/comments:** Add a comment to an existing post.

27. **GET /follow/posts/{postID}:** Get the data around and on a specific rss feed post. Works in tandem with the share functionality.

28. **GET /feeds/sample-posts/{feedID}:** Get sample random posts for specified posts to demonstrate a "taste" of what they look like.

29. **GET /top/creators:** Get the top x users of the API. it uses an algorithm explained in this readme.

30. **GET /search-options/feeds:** Get all available feeds to populate your search filter.

31. **GET /search-options/feed-type:** Get all available feed-type to populate your search filters.

32. **GET /subscriptions:** Get all transactional/subscriptional data for a specific users

33. **POST /subscriptions/initialize:** Initializes a subscription intent, which will return a redirect to the payment gateway

34. **POST /subscriptions/verify:** Verifies a transation made by a specific user via the gateway sent back from the init request

35. **GET /subscriptions/plans:** Gets  back all subscription plans supported by the app.

36. **POST /subscriptions/plans:** Add a subscription plan including details such as features, prices and more.

37. **GET /subscriptions/challenged:** A poll endpoint to check whether a user has a challenged subscription transaction.

38. **PATCH /subscriptions/challenged:** Update challenged transactions. Doing this will only delay a recurring charge

39. **PATCH /subscriptions:** Allows a user to cancel an existing subscription, preventing any further recurring charges.

40. **POST /api/activation:** Allows a manual request for a new Reset token email for new registered users 

<hr />

## 🔧 Running the tests <a name = "tests"></a>

The project has existing tests represented by files ending with the word `"_test"` e.g `rssdata_test.go`

### Break down into end to end tests

Each test file contains a myriad of tests to run on various entities mainly functions.
The test files are organized into `structs of tests` and their corresponding test logic.

You can run them directly from the vscode test UI. Below represents test results for the scraper:

```bash
=== RUN   Test_application_rssFeedScraper
=== RUN   Test_application_rssFeedScraper/Test1
Fetching:  bbc
--- PASS: Test_application_rssFeedScraper/Test1 (0.00s)
=== RUN   Test_application_rssFeedScraper/Test2
Fetching:  Lane's
--- PASS: Test_application_rssFeedScraper/Test2 (0.00s)
=== RUN   Test_application_rssFeedScraper/Test3
Fetching:  Megaphone
--- PASS: Test_application_rssFeedScraper/Test3 (0.00s)
=== RUN   Test_application_rssFeedScraper/Test4
Fetching:  Daily Podcast
--- PASS: Test_application_rssFeedScraper/Test4 (0.00s)
=== RUN   Test_application_rssFeedScraper/Test5
Fetching:  Endagadget
--- PASS: Test_application_rssFeedScraper/Test5 (0.00s)
--- PASS: Test_application_rssFeedScraper (0.00s)
PASS
ok      github.com/blue-davinci/aggregate/cmd/api       0.874s
```
- <b>All other tests follow a similar prologue.</b>

<hr />

## 🎈 Usage <a name="usage"></a>
As earlier mentioned, the api uses a myriad of flags which you can use to launch the application.
An example of launching the application with your `smtp server's setting` includes:
```bash
make build/api ## build api using the makefile
./bin/api.exe -smtp-username=pigbearman -smtp-password=algor ## run the built api with your own values

Direct Run: 
go run main.go
```

## 🧩 Algo <a name = "algo"></a>
We calculate the score for each user a bit differently. Although v1 was a simple feed follows and creation division, we moved and now the algorithm looks like this:

**score = (total feeds) / Σ(follows \* w_f \* e^(-λ \* t_f) + likes \* w_l \* e^(-λ \* t_l) + comments \* w_c \* e^(-λ \* t_c) + R) \* 100**

Where:

- **w_f** = weight for follows (e.g., 1.0)
- **w_l** = weight for likes (e.g., 0.5)
- **w_c** = weight for comments (e.g., 0.2)
- **t_f** = time since follow (in days)
- **t_l** = time since like (in days)
- **t_c** = time since comment (in days)
- **λ** = decay constant (controls how fast the weight decreases over time, e.g., 0.01)
- **R** = random factor for users with minimal activity to ensure they don't have identical scores (e.g., a small random value between 0 and 1)

### Example Calculation

**User X:**

- Total Follows: 100
- Total Likes: 50
- Total Comments: 20
- Total Created Feeds: 10
- Average Time Between Feeds: 30 days

**Engagement Score:**

**Engagement Score = (100 \* 0.7) + (50 \* 0.3) + (20 \* 0.2) = 70 + 15 + 4 = 89**

**Consistency Score:**

**Consistency Score = (10 / 30) = 0.33**

(normalize to 0-100 range: **0.33 \* 100 = 33**)

**Random Factor:**

**Random Factor = R (a small random value between 0 and 1)**

**Final Score:**

**Final Score = ((89 \* 0.8) + (33 \* 0.2) + R) = 71.2 + 6.6 + R = 77.8 + R**

---

- Feel free to update the weights and the decay constant as per the requirements. The random factor ensures that users with minimal activity don't end up with identical scores.
- This is the **v3** edition of the above algo.

## 💳 Payment <a name = "Payment"></a>
This application uses Pay-Stack to handle payments. For the setup, you will need to:
1. Register to [paystack](https://paystack.com)

2. Get the Paystack `API` and add the following to the `.env` file in the `cmd\api` dir as below:
```bash
PAYSTACK_SECRET_KEY=xxxx-paystack-api-xxxxxxx
```
3. That is all you need for the setup. The paystack API works on the basis of an initialization and verification which can be done via a `webhook` or `poll`.

4. As it works in tandem with the app's **Limitation** parameters, you can change the parameters by using the **limitation** `flags` already listed above.

**Please Note:** The application also supports payments through **Mobile Money** in addition to supported Cards.

## 🚀 Deployment <a name = "deployment"></a>

This application can be deployed using Docker and Docker Compose. Here are the steps to do so:
1. **Build the Docker image:** by navigating to the root directory.
```bash
cd aggregate
```
2. **Verify Configs:** Check and verify the following file incase you want to change any configs:
```bash
- docker-compose.yml
- Dockerfile
```
3. **Build The Container:** Run the following command to build the docker image based on the `docker-compose.yml` file:
```
docker compose up --build
```


Please remember you can use flags, mentioned [here](#optpars) while running the api by setting them in
the `Dockerfile` like so:

```bash
CMD ["./bin/api.exe", "-smtp-username", "smtp username", "-port", "your_port", "-smtp-password", "your_smtp_pass"]
```

<b>Note:</b> <strong>There is a pre-built</strong> package for anyone who may feel <i>less enthusiastic</i> about building
it themselves. You can get it going by doing:

```bash
docker pull ghcr.io/blue-davinci/aggregate:latest
```
<hr />

## ⛏️ Built Using <a name = "built_using"></a>
- [Go](https://golang.org/) - Backend
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Paystack](https://paystack.com) - Payment processing
- [SQLC](https://github.com/kyleconroy/sqlc) - Generate type safe Go from SQL
- [Goose](https://github.com/pressly/goose) - Database migration tool
- [HTML/CSS](https://developer.mozilla.org/en-US/docs/Web/HTML) - Templates

## ✍️ Authors <a name = "authors"></a>

- [@blue-davinci](https://github.com/blue-davinci) - Idea & Initial work

See also the list of [contributors](https://github.com/blue-davinci/aggregate/contributors) who participated in this project.

<hr />

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- Hat tip to anyone whose code was used

## 📰 Inspiration:
Aggregate was born from the necessity to streamline how we consume and manage the vast ocean of information available online. As avid tech enthusiasts and news followers, we often found ourselves juggling between multiple sources to stay updated. 

This sparked the idea to create a unified platform where all our favorite feeds could be aggregated into one seamless experience. Thus, Aggregate was conceived with a mission to simplify and enhance the way we access information.

## 📚 References <a name = "references"></a>

- [Go Documentation](https://golang.org/doc/): Official Go documentation and tutorials.
- [PostgreSQL Documentation](https://www.postgresql.org/docs/): Official PostgreSQL documentation.
- [SQLC Documentation](https://docs.sqlc.dev/en/latest/): Official SQLC documentation and guides.
