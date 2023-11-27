//go:build integration

package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"webapp/pkg/data"
	"webapp/pkg/repository"
)

var (
	host     = "localhost"
	user     = "web_user"
	password = "password"
	dbName   = "web"
	port     = "5432"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource

var pool *dockertest.Pool

var testDB *sql.DB

var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; is it running? %s", err)
	}

	pool = p

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource:%s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("Error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to database: %s", err)
	}

	err = createTables()
	if err != nil {
		log.Fatalf("error creating tables:%s", err)
	}

	testRepo = &PostgresDBRepo{
		DB: testDB,
	}
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Test_PingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("can't ping DB")
	}
}

func TestPostgresDBRepoInsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("insert return and error: %s", err)
	}
	if id != 1 {
		t.Errorf("insert user returned wrong id; expected 1, but got %d", id)
	}
}

func TestPostgresDBRepoAllUser(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("all users reports an error:%s", err)
	}

	if len(users) != 1 {
		t.Errorf("all users reports wrong size; expected 1, but got %d", len(users))
	}

	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin2@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("all users reports an error:%s", err)
	}

	if len(users) != 2 {
		t.Errorf("all users reports wrong size; expected 2, but got %d", len(users))
	}

}

func TestPostgresDBRepoGetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("error getting user by id:%s", err)
	}

	if user.Email != "admin@example.com" {
		t.Errorf(
			"wrong email returned by GetUser; expected admin@example.com but got %s",
			user.Email,
		)
	}
}

func TestPostgresDBRepoGetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("admin@example.com")
	if err != nil {
		t.Errorf("error getting user by email:%s", err)
	}

	if user.Email != "admin@example.com" {
		t.Errorf(
			"wrong email returned by GetUserByEmail; expected admin@example.com but got %s",
			user.Email,
		)
	}
}

func TestPostgresDBRepoUpdateUser(t *testing.T) {
	user, _ := testRepo.GetUserByEmail("admin@example.com")

	user.FirstName = "Mat"
	user.Email = "mat@email.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("error updating user %d:%s", user.ID, err)
	}

	user, _ = testRepo.GetUser(user.ID)
	if user.FirstName != "Mat" || user.Email != "mat@email.com" {
		t.Errorf(
			"expected updated record to have first name Mat and email mat@emai.com, but got %s %s",
			user.FirstName,
			user.Email,
		)
	}
}

func TestPostgresDBRepoDeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)

	if err != nil {
		t.Errorf("expected not error when delete user but got %s", err)
	}

	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Errorf("expected err when try retrieve deleted but got no err")
	}
}

func TestPostgresDBRepoResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "password")

	if err != nil {
		t.Errorf("error when setting user password %s", err)
	}

	user, _ := testRepo.GetUser(1)
	matches, err := user.PasswordMatches("password")
	if err != nil {
		t.Errorf("error when trying check password %s", err)
	}
	if !matches {
		t.Errorf("password should match with 'password', but does")
	}
}

func TestPostgresDBRepoInsertUserImage(t *testing.T) {
	var image data.UserImage

	image.UserID = 1
	image.FileName = "test.jpg"
	image.CreatedAt = time.Now()
	image.UpdatedAt = time.Now()

	newID, err := testRepo.InsertUserImage(image)
	if err != nil {
		t.Error("insert user image failed:", err)
	}
	if newID != 1 {
		t.Errorf("got wrong id for image;should be 1, but got %d", newID)
	}

	image.UserID = 100
	_, err = testRepo.InsertUserImage(image)

	if err == nil {
		t.Error("should be got error when try insert user image with none user id")
	}
}
