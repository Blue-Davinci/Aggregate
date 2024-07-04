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

---

<p align="center"> Aggregate [Back-End]: This is the full Golang backend code for the <strong>AGGREGATE Project</strong> 🚀🌐
<hr />
You are currently in the BackEnd section, To view the FrontEnd go [here](https://github.com/Blue-Davinci/Aggregate-FronteEnd)
    <br> 
</p>

## 📝 Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Deployment](#deployment)
- [Usage](#usage)
- [Built Using](#built_using)
- [TODO](../TODO.md)
- [Contributing](../CONTRIBUTING.md)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

<hr />
## 🧐 About <a name = "about"></a>

Aggregate is a content aggregation platform designed to streamline information consumption. Its purpose is to centralize feeds from various sources—such as RSS and Atom—into a unified stream. Users can effortlessly follow their favorite content, whether it’s news, blogs, or other updates. The project emphasizes efficiency, security, and a user-friendly experience, making it a valuable tool for staying informed in today’s fast-paced digital landscape. 🚀🌐

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

  ```bash
  make help
  ```
  - **Output:**
  ```
  make help
  Usage: 
  run/api            -  run the api application
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
- **smtp-sender:** Sets the sender for SMTP (email) communications. Default: "Groovy <no-reply@groovy.com>".
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
- **smtp-sender [string]:** SMTP sender (default "Groovy <no-reply@groovy.com>")
- **smtp-username [string]:** SMTP username (default "skunkhunt42")
- **baseurl [string]:** frontend url (default "http://localhost:5173")
- **activationurl [string]:** frontend activation url (default "http://localhost:5173/verify?token=")
- **passwordreseturl:** frontend password reset url (default "http://localhost:5173/reset?token=")

Using `make run`, will run the API with a default connection string located 
in `cmd\api\.env`. If you're using `powershell`, you need to load the values otherwise you will get
a `cannot load env file` error. Use the PS code below to load it or change the env variable:
```powershell
$env:GROOVY_DB_DSN=(Get-Content -Path .\cmd\api\.env | Where-Object { $_ -match "GROOVY_DB_DSN" } | ForEach-Object { $($_.Split("=", 2)[1]) })
```

Alternatively, in unix systems you can make a .envrc file and load it directly in the makefile by importing like so:
```makefile
include .envrc
```

A succesful run will output:
```bash
make run/api
'Running cmd/api...'
go run ./cmd/api
{"level":"INFO","time":"2024-07-04T15:56:16Z","message":"Loading Environment Variables","properties":{"DSN":"cmd\\api\\.env"}}
{"level":"INFO","time":"2024-07-04T15:56:16Z","message":"database connection pool established"}
{"level":"INFO","time":"2024-07-04T15:56:16Z","message":"Starting RSS Feed Scraper","properties":{"Client Timeout":"15","Interval":"40s","No of Go Routines":"5","No of Retries":"3"}}
{"level":"INFO","time":"2024-07-04T15:56:16Z","message":"starting server","properties":{"addr":":4000","env":"development"}}
```

<hr />
### API Endpoints <a name = "endpoints"></a>
Below are all the end points for the API and a high level description of what they do.

1. **GET /v1/healthcheck:** Checks the health of the application. Returns a 200 OK status code if the application is running correctly.
2. **POST /v1/users:** Registers a new user.

3. **PUT /v1/users/activated:** Activates a user.

4. **POST /v1/api/authentication:** Creates an authentication token.

5. **GET /debug/vars:** Provides debug variables from the `expvar` package. 

6. **POST /feeds:** Add an RSS Type feed {Atom/RSS}

7. **GET /feeds?page=1&page_size=30:** Get all Feeds in the DB, <b>With Pagination</b>
    <b>Note:</b> <i>You can leave the pagination parameters foe default values!</i>

8. **POST /feeds/follows:** Follow any feed for a user

9. **GET /feeds/follow:** Get all feeds followed by a user

10. **DELETE /feeds/follow/{feed_id}:** Unfollow a followed feed

11. **GET /feeds:** Get all Posts from scraped feeds that are followed by a user.
-------------------------------------------------------------------------------------------
- **Password Reset:** To Be added, for password reset request
- **Manual Token request:** To Be added, for password manual token activation request
- **Notifications:** To Be Added, for user based notifications.

## 🔧 Running the tests <a name = "tests"></a>

The project has existing tests represented by files ending with the word `"_test"` e.g `rssdata_test.go`

### Break down into end to end tests

Each test file contains a myriad of tests to run on various entities mainly functions.
The test files are organized into `structs of tests` and their corresponding test logic.

You can run them directly from the vscode test UI. Below represents test results for the scraper:
```
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

## 🚀 Deployment <a name = "deployment"></a>

This application can be deployed using Docker and Docker Compose. Here are the steps to do so:
<strong>[TO BE ADDED]</strong>

<hr />

## ⛏️ Built Using <a name = "built_using"></a>
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Go](https://golang.org/) - Backend
- [HTML](https://developer.mozilla.org/en-US/docs/Web/HTML) - Email Template
- [JavaScript](https://developer.mozilla.org/en-US/docs/Web/JavaScript) - Frontend behavior

## ✍️ Authors <a name = "authors"></a>

- [@blue-davinci](https://github.com/blue-davinci) - Idea & Initial work

See also the list of [contributors](https://github.com/blue-davinci/aggregate/contributors) who participated in this project.

<hr />

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- Hat tip to anyone whose code was used

- Inspiration

## 📚 References <a name = "references"></a>

- [Go Documentation](https://golang.org/doc/): Official Go documentation and tutorials.
- [PostgreSQL Documentation](https://www.postgresql.org/docs/): Official PostgreSQL documentation.
