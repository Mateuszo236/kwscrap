package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// Ksiega represents a property registry record
type Ksiega struct {
	KodSadu        string // Court code
	Numer          string // Registry number
	CyfraKontrolna string // Control digit
}

// randomSleep returns a chromedp action that sleeps for a random duration between min and max seconds
// This simulates human-like behavior to avoid detection
func randomSleep(min, max float64) chromedp.Action {
	return chromedp.ActionFunc(func(context.Context) error {
		duration := min + rand.Float64()*(max-min)
		time.Sleep(time.Duration(duration * float64(time.Second)))
		return nil
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	/*mojeKsiegi := []Ksiega{
		{"OL1O", "00140441", "9", "10/1"},
	}*/

	// Konfiguracja ChromeDP z ukryciem automatyzacji
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-maximized", true),

		// Masking automation detection
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("use-mock-keychain", false),

		// Polish language and user agent for realistic requests
		chromedp.Flag("lang", "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	fmt.Println("Starting scraping...")

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Inject script to hide webdriver property before page load
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(c context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(`
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		`).Do(c)
		return err
	})); err != nil {
		log.Fatal("Failed to hide webdriver:", err)
	}

	// Set timeout for long-running operations
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	fmt.Println("Starting in stealth mode...")

	// Iterate through generated registry numbers
	licznik := 0
	for start := 104; start <= 140450; start++ {
		numerFormatted := fmt.Sprintf("%08d", start)
		cyfraKontrolna, err := ObliczCyfreKontrolna("OL1O", numerFormatted)
		if err != nil {
			log.Printf("Error calculating control digit: %v", err)
			continue
		}

		kw := Ksiega{
			KodSadu:        "OL1O",
			Numer:          numerFormatted,
			CyfraKontrolna: cyfraKontrolna,
		}

		licznik++
		fmt.Printf("[%d/%d] Processing: %s/%s/%s\n", licznik, 140450-104, kw.KodSadu, kw.Numer, kw.CyfraKontrolna)
		var htmlContent string
		var searchResultsHTML string

		err = chromedp.Run(ctx,
			// Navigate to the registry search page
			chromedp.Navigate(`https://przegladarka-ekw.ms.gov.pl/eukw_prz/KsiegiWieczyste/wyszukiwanieKW`),

			// Wait for search form to load
			chromedp.WaitVisible(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(0.3, 0.7),

			// 1. Pole Sądu
			chromedp.Click(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(0.2, 0.4),

			// Czyścimy pole
			chromedp.KeyEvent(kb.Control+"a"),
			chromedp.KeyEvent(kb.Backspace),

			chromedp.SendKeys(`#kodWydzialuInput`, kw.KodSadu, chromedp.ByID),
			randomSleep(0.3, 0.7),

			chromedp.KeyEvent(kb.ArrowDown),
			randomSleep(0.2, 0.4),
			chromedp.KeyEvent(kb.Enter),

			// 2. Czekamy na AJAX
			randomSleep(1.0, 2.0),

			// 3. Reset Focusa
			chromedp.Click(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(0.3, 0.6),

			// 4. Pole Numeru
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.2, 0.4),
			chromedp.KeyEvent(kw.Numer),

			// Fill in control digit
			randomSleep(0.2, 0.4),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.2, 0.4),
			chromedp.KeyEvent(kw.CyfraKontrolna),

			// Tab to search button and submit
			chromedp.KeyEvent(kb.Enter),

			// Wait for search results to load
			randomSleep(0.3, 0.6),

			// Check if registry exists; abort if not found
			chromedp.Evaluate(`document.documentElement.outerHTML`, &searchResultsHTML),
			chromedp.ActionFunc(func(c context.Context) error {
				if strings.Contains(searchResultsHTML, "nie została odnaleziona") {
					fmt.Printf("Registry %s/%s/%s not found, skipping.\n", kw.KodSadu, kw.Numer, kw.CyfraKontrolna)
					return fmt.Errorf("not found")
				}
				return nil
			}),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.01, 0.03),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.2, 0.4),
			chromedp.KeyEvent(kb.Tab),

			// Open registry details page
			randomSleep(1.0, 2.0),
			chromedp.KeyEvent(kb.Enter),

			// Wait for page load
			randomSleep(2.0, 3.0),

			// Extract full HTML content
			chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlContent),
		)

		if err != nil {
			log.Printf("Error during registry processing: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Save HTML to file
		filename := fmt.Sprintf("output/%s_%s_%s.html", kw.KodSadu, kw.Numer, kw.CyfraKontrolna)

		if err := os.MkdirAll("output", 0755); err != nil {
			log.Printf("Failed to create output directory: %v", err)
			continue
		}

		if err := os.WriteFile(filename, []byte(htmlContent), 0644); err != nil {
			log.Printf("Failed to write file %s: %v", filename, err)
		} else {
			fmt.Printf("Saved: %s\n", filename)
		}

		// Rate limiting: random delay between requests
		wait := 1 + rand.Intn(3)
		time.Sleep(time.Duration(wait) * time.Second)
	}
	fmt.Println("Done.")
}
