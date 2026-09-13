// Command plans manages the entities.Plan catalog and assigns plans to users.
// It is an admin-only tool run via the container shell (no HTTP surface):
//
//	docker exec httpsms-api go run ./cmd/plans seed [--file=cmd/plans/plans.json]
//	docker exec httpsms-api go run ./cmd/plans list
//	docker exec httpsms-api go run ./cmd/plans assign --email=user@example.com --plan=pro-monthly
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/NdoleStudio/httpsms/pkg/di"
	"github.com/NdoleStudio/httpsms/pkg/entities"
)

// planCatalogEntry mirrors the shape of one entry in the JSON plan catalog file
type planCatalogEntry struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	MessageLimit  int    `json:"message_limit"`
	PhoneLimit    int    `json:"phone_limit"`
	PriceCents    int    `json:"price_cents"`
	Currency      string `json:"currency"`
	BillingPeriod string `json:"billing_period"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	container := di.NewContainer(os.Getenv("GCP_PROJECT_ID"), "")
	ctx := context.Background()

	switch os.Args[1] {
	case "seed":
		seedCmd(ctx, container)
	case "list":
		listCmd(ctx, container)
	case "assign":
		assignCmd(ctx, container)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Usage:
  go run ./cmd/plans seed [--file=cmd/plans/plans.json]        Idempotently create/update plans from a JSON catalog
  go run ./cmd/plans list                                      List all plans currently in the database
  go run ./cmd/plans assign --email=<email> --plan=<name>      Assign a plan to a user`)
}

func seedCmd(ctx context.Context, container *di.Container) {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	file := fs.String("file", "cmd/plans/plans.json", "path to the JSON plan catalog")
	_ = fs.Parse(os.Args[2:])

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Printf("cannot read file [%s]: %v\n", *file, err)
		os.Exit(1)
	}

	var entries []planCatalogEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("cannot parse JSON in [%s]: %v\n", *file, err)
		os.Exit(1)
	}

	repository := container.PlanRepository()
	for _, entry := range entries {
		plan := &entities.Plan{
			Name:          entry.Name,
			DisplayName:   entry.DisplayName,
			MessageLimit:  entry.MessageLimit,
			PhoneLimit:    entry.PhoneLimit,
			PriceCents:    entry.PriceCents,
			Currency:      entry.Currency,
			BillingPeriod: entry.BillingPeriod,
		}
		if err = repository.Upsert(ctx, plan); err != nil {
			fmt.Printf("cannot upsert plan [%s]: %v\n", entry.Name, err)
			os.Exit(1)
		}
		fmt.Printf("upserted plan [%s] (messages=%d, phones=%d, price=%d %s/%s)\n", entry.Name, entry.MessageLimit, entry.PhoneLimit, entry.PriceCents, entry.Currency, entry.BillingPeriod)
	}

	fmt.Printf("done: %d plan(s) seeded from [%s]\n", len(entries), *file)
}

func listCmd(ctx context.Context, container *di.Container) {
	plans, err := container.PlanRepository().Index(ctx)
	if err != nil {
		fmt.Printf("cannot list plans: %v\n", err)
		os.Exit(1)
	}

	for _, plan := range plans {
		fmt.Printf("%-16s %-20s messages=%-8d phones=%-4d price=%d %s/%s\n", plan.Name, plan.DisplayName, plan.MessageLimit, plan.PhoneLimit, plan.PriceCents, plan.Currency, plan.BillingPeriod)
	}
}

func assignCmd(ctx context.Context, container *di.Container) {
	fs := flag.NewFlagSet("assign", flag.ExitOnError)
	email := fs.String("email", "", "email of the user")
	planName := fs.String("plan", "", "name of the plan to assign (must already exist - run 'seed' first)")
	_ = fs.Parse(os.Args[2:])

	if *email == "" || *planName == "" {
		fmt.Println("--email and --plan are required")
		os.Exit(1)
	}

	planRepository := container.PlanRepository()
	if _, err := planRepository.LoadByName(ctx, *planName); err != nil {
		fmt.Printf("plan [%s] does not exist - run 'go run ./cmd/plans seed' first\n", *planName)
		os.Exit(1)
	}

	userRepository := container.UserRepository()
	user, err := userRepository.LoadByEmail(ctx, *email)
	if err != nil {
		fmt.Printf("cannot find user with email [%s]: %v\n", *email, err)
		os.Exit(1)
	}

	previousPlan := user.SubscriptionName
	user.SubscriptionName = entities.SubscriptionName(*planName)
	if err = userRepository.Update(ctx, user); err != nil {
		fmt.Printf("cannot update user [%s]: %v\n", *email, err)
		os.Exit(1)
	}

	fmt.Printf("assigned plan [%s] to user [%s] (%s) - was [%s]\n", *planName, user.Email, user.ID, previousPlan)
}
