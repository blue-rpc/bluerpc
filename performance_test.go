package bluerpc

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

type performance_test_input struct {
	House string `paramName:"house" validate:"required"`
}
type performance_test_output struct {
	FieldOneOut   string   `paramName:"fieldOneOut" validate:"required"`
	FieldTwoOut   string   `paramName:"fieldTwoOut" `
	FieldThreeOut string   `paramName:"fieldThreeOut" validate:"required"`
	FieldFourOut  []string `paramName:"fieldFourOut" `
}

func TestPerf10(t *testing.T) {
	fmt.Printf(DefaultColors.Green + "TESTING PERFORMANCE OF 10 DEEP NESTED PROCEDURES: \n" + DefaultColors.Reset)

	validate := validator.New(validator.WithRequiredStructEnabled())
	app := New(&Config{
		OutputPath:          "./local-some-file.ts",
		ValidatorFn:         validate.Struct,
		DisableInfoPrinting: true,
	})

	var wg sync.WaitGroup
	fmt.Printf(DefaultColors.Green + "DEPTH: 10 " + DefaultColors.Reset)
	avgTenTime := getAvg(func() time.Duration {
		return perfLoop(app, 10, &wg)
	}, &wg)
	fmt.Printf(DefaultColors.Green+"AVERAGE TIME FOR GENERATING DEPTH OF 10: %s\n"+DefaultColors.Reset, avgTenTime)

}

func TestPerf100(t *testing.T) {
	fmt.Printf(DefaultColors.Green + "TESTING PERFORMANCE OF 100 DEEP NESTED PROCEDURES: \n" + DefaultColors.Reset)

	validate := validator.New(validator.WithRequiredStructEnabled())
	app := New(&Config{
		OutputPath:          "./local-some-file.ts",
		ValidatorFn:         validate.Struct,
		DisableInfoPrinting: true,
	})

	var wg sync.WaitGroup
	fmt.Printf(DefaultColors.Green + "DEPTH: 100 " + DefaultColors.Reset)
	avgHundredTime := getAvg(func() time.Duration {
		return perfLoop(app, 100, &wg)
	}, &wg)
	fmt.Printf(DefaultColors.Green+"AVERAGE TIME FOR GENERATING DEPTH OF 100: %s\n"+DefaultColors.Reset, avgHundredTime)

}
func TestPerf1000(t *testing.T) {
	fmt.Printf(DefaultColors.Green + "TESTING PERFORMANCE OF 1000 DEEP NESTED PROCEDURES: \n" + DefaultColors.Reset)

	validate := validator.New(validator.WithRequiredStructEnabled())
	app := New(&Config{
		OutputPath:          "./local-some-file.ts",
		ValidatorFn:         validate.Struct,
		DisableInfoPrinting: true,
	})

	var wg sync.WaitGroup
	fmt.Printf(DefaultColors.Green + "DEPTH: 1000 " + DefaultColors.Reset)
	avgHundredTime := getAvg(func() time.Duration {
		return perfLoop(app, 100, &wg)
	}, &wg)
	fmt.Printf(DefaultColors.Green+"AVERAGE TIME FOR GENERATING DEPTH OF 1000: %s\n"+DefaultColors.Reset, avgHundredTime)

}
func perfLoop(app *App, num int, wg *sync.WaitGroup) time.Duration {
	wg.Add(1)
	defer wg.Done()
	query := NewQuery[test_query, performance_test_output](app, func(ctx *Ctx, query test_query) (*Res[performance_test_output], error) {
		return &Res[performance_test_output]{
			Body: performance_test_output{
				FieldOneOut:   "dwa",
				FieldTwoOut:   "dwadwa",
				FieldThreeOut: "dwadwadwa",
			},
		}, nil
	})
	mut := NewMutation[test_query, performance_test_input, performance_test_output](app, func(ctx *Ctx, query test_query, input performance_test_input) (*Res[performance_test_output], error) {
		return &Res[performance_test_output]{
			Body: performance_test_output{
				FieldOneOut:   "dwadwa",
				FieldTwoOut:   "dwadwadwa",
				FieldThreeOut: "dwadwadwad",
			},
		}, nil
	})
	fmt.Println("num of rounds", num)
	currGroup := app.Router("/start")
	for i := 0; i < num; i++ {

		newGrp := currGroup.Router(fmt.Sprintf("depth%d", i))
		query.Attach(newGrp, "/query")
		mut.Attach(newGrp, "/mutation")

	}
	start := time.Now()

	go func() {
		err := app.Listen(fmt.Sprintf(":%d", 3000))
		if err != nil {
			fmt.Println("Error starting the server:", err)
			wg.Done()
			return
		}

	}()

	for {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d", 3000))
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond) // Adjust the sleep duration as needed
	}

	// Record elapsed time
	elapsed := time.Since(start)

	// Shut down the server
	app.Shutdown()

	return elapsed
}
func getAvg(someFunc func() time.Duration, wg *sync.WaitGroup) time.Duration {
	var total time.Duration

	// Run the function 100 times
	for i := 0; i < 10; i++ {
		// Measure the time taken by the function
		duration := someFunc()
		total += duration
		wg.Wait()

	}

	// Calculate the average time
	avg := total / 10
	return avg

}
