package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/vkalekis/companies/internal/cdc"
	"github.com/vkalekis/companies/internal/config"
	"github.com/vkalekis/companies/pkg/model"
)

var (
	db       = "company_db"
	username = "company_admin"
	password = "company_admin_pw"

	serverPort = 8080
)

func TestApp(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:18.3-alpine",
		postgres.WithDatabase(db),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("error initializing postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("error terminating postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(
		&config.Config{
			Database: struct {
				Host     string `yaml:"host"`
				Port     int    `yaml:"port"`
				Username string `yaml:"username"`
				Password string `yaml:"password"`
				DBName   string `yaml:"dbName"`
			}{
				Host:     host,
				Port:     port.Int(),
				Username: username,
				Password: password,
				DBName:   db,
			},
			Server: struct {
				Port int `yaml:"port"`
			}{
				Port: serverPort,
			},
			CDC: struct {
				Operator string `yaml:"operator"`
			}{
				Operator: "log",
			},
			Kafka: struct {
				BootstrapServers string `yaml:"bootstrapServers"`
				Topic            string `yaml:"topic"`
			}{
				BootstrapServers: "",
				Topic:            "",
			},
		},
		log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Inject the test CDCOperator
	cdcOperator := &collectorCDCOperator{}
	app.cdcOperator = cdcOperator

	if err := app.Start(); err != nil {
		t.Fatal(err)
	}
	defer app.Stop()

	// Wait for the app to start
	time.Sleep(2 * time.Second)

	client := http.Client{}

	// ------------------------------------------------------------
	// Login
	// ------------------------------------------------------------

	data, _ := json.Marshal(model.Login{
		Username: "user",
		Password: "pass",
	})

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/login", host, serverPort), bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	var jwt model.JWT
	err = json.Unmarshal(respBody, &jwt)
	if err != nil {
		t.Fatal(err)
	}

	if len(jwt.JWT) == 0 {
		t.Fatal("empty jtw")
	}

	// ------------------------------------------------------------
	// Create a dummy company
	// ------------------------------------------------------------

	var description = "My company"

	data, _ = json.Marshal(model.CompanyCreateReq{
		Name:        "Test",
		Description: &description,
		Employees:   12,
		Registered:  true,
		CompanyType: model.CompanyType_Cooperative,
	})

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/companies", host, serverPort), bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt.JWT))

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	var company model.Company
	err = json.Unmarshal(respBody, &company)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Company=%v", company)

	if company.Name != "Test" ||
		(company.Description != nil && description != "My company") ||
		company.Employees != 12 ||
		company.Registered != true ||
		company.CompanyType != model.CompanyType_Cooperative {
		t.Fatalf("fields do not match")
	}

	var id = company.Id

	// ------------------------------------------------------------
	// Fetch it from the BE
	// ------------------------------------------------------------

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/companies/%s", host, serverPort, id), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	err = json.Unmarshal(respBody, &company)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Company=%v", company)

	if company.Id != id ||
		company.Name != "Test" ||
		(company.Description != nil && description != "My company") ||
		company.Employees != 12 ||
		company.Registered != true ||
		company.CompanyType != model.CompanyType_Cooperative {
		t.Fatalf("fields do not match")
	}

	// ------------------------------------------------------------
	// Do an update
	// ------------------------------------------------------------

	employees := 50
	companyType := model.CompanyType(model.CompanyType_NonProfit)
	data, _ = json.Marshal(model.CompanyUpdateReq{
		Employees:   &employees,
		CompanyType: &companyType,
	})

	req, err = http.NewRequest(http.MethodPatch, fmt.Sprintf("http://%s:%d/companies/%s", host, serverPort, id), bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt.JWT))

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	err = json.Unmarshal(respBody, &company)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(company)
	if company.Id != id ||
		company.Name != "Test" ||
		(company.Description != nil && description != "My company") ||
		company.Employees != 50 ||
		company.Registered != true ||
		company.CompanyType != model.CompanyType_NonProfit {
		t.Fatalf("fields do not match")
	}

	// ------------------------------------------------------------
	// Fetch it from the BE (again)
	// ------------------------------------------------------------

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/companies/%s", host, serverPort, id), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	err = json.Unmarshal(respBody, &company)
	if err != nil {
		t.Fatal(err)
	}

	if company.Id != id ||
		company.Name != "Test" ||
		(company.Description != nil && description != "My company") ||
		company.Employees != 50 ||
		company.Registered != true ||
		company.CompanyType != model.CompanyType_NonProfit {
		t.Fatalf("fields do not match")
	}

	// ------------------------------------------------------------
	// Delete it
	// ------------------------------------------------------------

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/companies/%s", host, serverPort, id), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt.JWT))

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("unexpected http status code")
	}

	// ------------------------------------------------------------
	// Try to fetch it from the BE
	// ------------------------------------------------------------

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/companies/%s", host, serverPort, id), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("unexpected http status code")
	}

	// ------------------------------------------------------------
	// Create an invalid company
	// ------------------------------------------------------------

	data, _ = json.Marshal(model.Company{
		Name:        "Test",
		Description: &description,
		Employees:   12,
		Registered:  true,
		CompanyType: "INVALID",
	})

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/companies", host, serverPort), bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt.JWT))

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("unexpected http status code")
	}

	// ------------------------------------------------------------
	// Create an company without auth first
	// ------------------------------------------------------------

	data, _ = json.Marshal(model.Company{
		Name:        "Test",
		Description: &description,
		Employees:   12,
		Registered:  true,
		CompanyType: "INVALID",
	})

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/companies", host, serverPort), bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// ---> req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt.JWT))

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatal("unexpected http status code")
	}

	// ------------------------------------------------------------
	// Check for the collected CDC operations
	// ------------------------------------------------------------

	wantCDCOps := []cdc.Operation{
		{
			Before: nil,
			After: &model.Company{
				Id:          id,
				Name:        "Test",
				Description: &description,
				Employees:   12,
				Registered:  true,
				CompanyType: model.CompanyType_Cooperative,
			},
			Op: cdc.Op_Create,
		},
		{
			Before: &model.Company{
				Id:          id,
				Name:        "Test",
				Description: &description,
				Employees:   12,
				Registered:  true,
				CompanyType: model.CompanyType_Cooperative,
			},
			After: &model.Company{
				Id:          id,
				Name:        "Test",
				Description: &description,
				Employees:   50,
				Registered:  true,
				CompanyType: model.CompanyType_NonProfit,
			},
			Op: cdc.Op_Update,
		},
		{
			Before: &model.Company{
				Id:          id,
				Name:        "Test",
				Description: &description,
				Employees:   50,
				Registered:  true,
				CompanyType: model.CompanyType_NonProfit,
			},
			After: nil,
			Op:    cdc.Op_Delete,
		},
	}
	if !areCDCOpsEqual(cdcOperator.ops, wantCDCOps) {
		t.Error("error matching cdc ops")
		t.Logf("got: \n")
		for _, op := range cdcOperator.ops {
			t.Logf("\t op=%s, bef=%+v, after=%+v", op.Op, op.Before, op.After)
		}
		t.Logf("want: \n")
		for _, op := range wantCDCOps {
			t.Logf("\t op=%s, bef=%+v, after=%+v", op.Op, op.Before, op.After)
		}
	}
}

type collectorCDCOperator struct {
	ops []cdc.Operation
}

func (operator *collectorCDCOperator) LogCDCOperation(op cdc.Operation) {
	operator.ops = append(operator.ops, op)
}

func areCDCOpsEqual(a, b []cdc.Operation) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range len(a) {
		if a[i].Op != b[i].Op {
			return false
		}

		if (a[i].Before == nil && b[i].Before != nil) || (a[i].Before != nil && b[i].Before == nil) {
			return false
		}

		if (a[i].After == nil && b[i].After != nil) || (a[i].After != nil && b[i].After == nil) {
			return false
		}

		if a[i].Before != nil {
			// Disregard the ts
			a[i].Before.CreatedAt = time.Time{}
			a[i].Before.UpdatedAt = time.Time{}
			b[i].Before.CreatedAt = time.Time{}
			b[i].Before.UpdatedAt = time.Time{}
			if !reflect.DeepEqual(*a[i].Before, *b[i].Before) {
				return false
			}
		}

		if a[i].After != nil {
			// Disregard the ts
			a[i].After.CreatedAt = time.Time{}
			a[i].After.UpdatedAt = time.Time{}
			b[i].After.CreatedAt = time.Time{}
			b[i].After.UpdatedAt = time.Time{}
			if !reflect.DeepEqual(*a[i].After, *b[i].After) {
				return false
			}
		}
	}

	return true
}
