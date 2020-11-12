package postgres

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"database/sql"

	// import used by database/sql
	_ "gopkg.in/go-on/pq.v2"

	"github.com/edgedb/webapp-bench/_go/bench"
	"github.com/edgedb/webapp-bench/_go/cli"
)

func PQWorker(args cli.Args) (bench.Exec, bench.Close) {
	db, err := sql.Open("postgres", "user=postgres_bench dbname=postgres_bench password=edgedbbenchmark")
	if err != nil {
		log.Fatal(err)
	}

	// the get movie query uses 4 at a time
	db.SetMaxIdleConns(4 * args.Concurrency)

	regex := regexp.MustCompile(`users|movie|person`)
	queryType := regex.FindString(args.Query)

	var exec bench.Exec

	switch queryType {
	case "movie":
		exec = pqExecMovie(db, args)
	case "person":
		exec = pqExecPerson(db, args)
	case "users":
		exec = pqExecUser(db, args)
	default:
		log.Fatalf("unknown query type: %q", queryType)
	}

	close := func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	return exec, close
}

func pqExecMovie(db *sql.DB, args cli.Args) bench.Exec {
	var (
		movie Movie
		wg    sync.WaitGroup
	)

	queries := strings.Split(args.Query, ";")
	m := pqGetMovie(db, queries[0], &movie)
	d := pqGetDirectors(db, queries[1], &movie)
	c := pqGetCast(db, queries[2], &movie)
	r := pqGetReviews(db, queries[3], &movie)

	return func(id string) (time.Duration, string) {
		wg.Add(4)
		start := time.Now()
		go m(id, &wg)
		go d(id, &wg)
		go c(id, &wg)
		go r(id, &wg)
		wg.Wait()

		serial, err := json.Marshal(movie)
		if err != nil {
			log.Fatal(err)
		}

		duration := time.Since(start)
		return duration, string(serial)
	}
}

func pqGetMovie(
	db *sql.DB,
	query string,
	movie *Movie,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		row := stmt.QueryRow(id)
		row.Scan(
			&movie.ID,
			&movie.Image,
			&movie.Title,
			&movie.Year,
			&movie.Description,
			&movie.AvgRating,
		)

		err := row.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqGetDirectors(
	db *sql.DB,
	query string,
	movie *Movie,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var person MovieQueryPerson

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		movie.Directors = movie.Directors[:0]
		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(
				&person.ID,
				&person.FullName,
				&person.Image,
			)
			movie.Directors = append(movie.Directors, person)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqGetCast(
	db *sql.DB,
	query string,
	movie *Movie,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var person MovieQueryPerson

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		movie.Cast = movie.Cast[:0]
		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(
				&person.ID,
				&person.FullName,
				&person.Image,
			)
			movie.Cast = append(movie.Cast, person)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqGetReviews(
	db *sql.DB,
	query string,
	movie *Movie,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var review MovieQueryReview

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		movie.Reviews = movie.Reviews[:0]
		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(
				&review.ID,
				&review.Body,
				&review.Rating,
				&review.Author.ID,
				&review.Author.Name,
				&review.Author.Image,
			)
			movie.Reviews = append(movie.Reviews, review)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqExecPerson(db *sql.DB, args cli.Args) bench.Exec {
	var (
		person Person
		wg     sync.WaitGroup
	)

	queries := strings.Split(args.Query, ";")
	p := pqGetPerson(db, queries[0], &person)
	a := pqGetActedIn(db, queries[1], &person)
	d := pqGetDirected(db, queries[2], &person)

	return func(id string) (time.Duration, string) {
		wg.Add(3)
		start := time.Now()
		go p(id, &wg)
		go a(id, &wg)
		go d(id, &wg)
		wg.Wait()

		serial, err := json.Marshal(person)
		if err != nil {
			log.Fatal(err)
		}

		duration := time.Since(start)
		return duration, string(serial)
	}
}

func pqGetPerson(
	db *sql.DB,
	query string,
	person *Person,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		row := stmt.QueryRow(id)
		row.Scan(
			&person.ID,
			&person.FullName,
			&person.Image,
			&person.Bio,
		)

		err := row.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqGetActedIn(
	db *sql.DB,
	query string,
	person *Person,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var movie PersonQueryMovie

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		person.ActedIn = person.ActedIn[:0]
		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(
				&movie.ID,
				&movie.Image,
				&movie.Title,
				&movie.Year,
				&movie.AvgRating,
			)
			person.ActedIn = append(person.ActedIn, movie)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqGetDirected(
	db *sql.DB,
	query string,
	person *Person,
) func(string, *sync.WaitGroup) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var movie PersonQueryMovie

	return func(id string, wg *sync.WaitGroup) {
		defer wg.Done()

		person.Directed = person.Directed[:0]
		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			rows.Scan(
				&movie.ID,
				&movie.Image,
				&movie.Title,
				&movie.Year,
				&movie.AvgRating,
			)
			person.Directed = append(person.Directed, movie)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pqExecUser(db *sql.DB, args cli.Args) bench.Exec {
	var (
		user   User
		review UserQueryReview
	)

	stmt, err := db.Prepare(args.Query)
	if err != nil {
		log.Fatal(err)
	}

	return func(id string) (time.Duration, string) {
		start := time.Now()

		rows, err := stmt.Query(id)
		if err != nil {
			log.Fatal(err)
		}

		user.LatestReviews = user.LatestReviews[:0]
		for rows.Next() {
			rows.Scan(
				&user.ID,
				&user.Name,
				&user.Image,
				&review.ID,
				&review.Body,
				&review.Rating,
				&review.Movie.ID,
				&review.Movie.Image,
				&review.Movie.Title,
				&review.Movie.AvgRating,
			)

			user.LatestReviews = append(user.LatestReviews, review)
		}

		serial, err := json.Marshal(user)
		if err != nil {
			log.Fatal(err)
		}

		duration := time.Since(start)
		return duration, string(serial)
	}
}