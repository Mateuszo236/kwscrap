package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

type Ksiega struct {
	KodSadu        string
	Numer          string
	CyfraKontrolna string
	OczekiwaneID   string
}

// Funkcja losująca czas oczekiwania (dla ludzkiego zachowania)
func randomSleep(min, max float64) chromedp.Action {
	return chromedp.ActionFunc(func(context.Context) error {
		duration := min + rand.Float64()*(max-min)
		time.Sleep(time.Duration(duration * float64(time.Second)))
		return nil
	})
}

func main() {
	// Inicjalizacja losowości
	rand.Seed(time.Now().UnixNano())

	mojeKsiegi := []Ksiega{
		{"OL1O", "00140441", "9", "10/1"},
	}

	// --- POPRAWIONA KONFIGURACJA (Bezpieczne flagi) ---
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-maximized", true),

		// 1. Ukrywamy flagi automatyzacji (wersja uproszczona, która nie powoduje błędu)
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("excludeSwitches", "enable-automation"), // Tutaj była zmiana z []string na string
		chromedp.Flag("use-mock-keychain", false),

		// 2. Ustawiamy język polski
		chromedp.Flag("lang", "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7"),

		// 3. User-Agent
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// --- MAGIA: Usuwamy ślad robota w JS ---
	// To musi być wykonane ZANIM załaduje się strona
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(c context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(`
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		`).Do(c)
		return err
	})); err != nil {
		log.Fatal("Błąd przy ukrywaniu webdrivera:", err)
	}

	// Timeout
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	fmt.Println("Startuję w trybie STEALTH (Pełne udawanie człowieka)...")

	for i, kw := range mojeKsiegi {
		fmt.Printf("\n[%d/%d] Sprawdzam: %s/%s/%s\n", i+1, len(mojeKsiegi), kw.KodSadu, kw.Numer, kw.CyfraKontrolna)
		var htmlContent string

		err := chromedp.Run(ctx,
			// 0. Wejście na stronę
			chromedp.Navigate(`https://przegladarka-ekw.ms.gov.pl/eukw_prz/KsiegiWieczyste/wyszukiwanieKW`),

			// Jeśli IP jest zbanowane, tu może pojawić się błąd lub strona błędu
			chromedp.WaitVisible(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(1.5, 3.0),

			// 1. Pole Sądu
			chromedp.Click(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(0.5, 1.0),

			// Czyścimy pole
			chromedp.KeyEvent(kb.Control+"a"),
			chromedp.KeyEvent(kb.Backspace),

			chromedp.SendKeys(`#kodWydzialuInput`, kw.KodSadu, chromedp.ByID),
			randomSleep(1.5, 2.5),

			chromedp.KeyEvent(kb.ArrowDown),
			randomSleep(0.5, 1.0),
			chromedp.KeyEvent(kb.Enter),

			// 2. Czekamy na AJAX
			randomSleep(4.0, 6.0),

			// 3. Reset Focusa
			chromedp.Click(`#kodWydzialuInput`, chromedp.ByID),
			randomSleep(0.8, 1.5),

			// 4. Pole Numeru
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.8, 1.5),
			chromedp.KeyEvent(kw.Numer),

			// 5. Pole Cyfry
			randomSleep(0.8, 1.5),
			chromedp.KeyEvent(kb.Tab),
			randomSleep(0.8, 1.5),
			chromedp.KeyEvent(kw.CyfraKontrolna),

			// 6. Kliknij Wyszukaj
			randomSleep(1.0, 2.0),
			chromedp.Click(`#wyszukaj`, chromedp.ByID),

			// Czekamy na wynik
			randomSleep(2.0, 4.0),

			// 7. Przejście do wyniku
			chromedp.WaitVisible(`input[value="Przeglądanie aktualnej treści KW"]`, chromedp.ByQuery),
			randomSleep(1.0, 2.0),
			chromedp.Click(`input[value="Przeglądanie aktualnej treści KW"]`, chromedp.ByQuery),

			// 8. Wejście w Dział I-O
			randomSleep(2.0, 4.0),
			chromedp.WaitVisible(`//a[contains(text(), "Dział I-O")]`, chromedp.BySearch),
			randomSleep(1.0, 2.0),
			chromedp.Click(`//a[contains(text(), "Dział I-O")]`, chromedp.BySearch),

			// 9. Pobranie danych
			chromedp.WaitVisible(`#contentDzialu`, chromedp.ByID),
			randomSleep(1.0, 2.0),
			chromedp.OuterHTML(`body`, &htmlContent),
		)

		if err != nil {
			log.Printf("Błąd: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if strings.Contains(htmlContent, kw.OczekiwaneID) {
			fmt.Printf("-> WYNIK: Działka %s JEST w księdze.\n", kw.OczekiwaneID)
		} else {
			fmt.Printf("-> WYNIK: Brak działki %s.\n", kw.OczekiwaneID)
		}

		// PRZERWA MIĘDZY KSIĘGAMI
		if i < len(mojeKsiegi)-1 {
			wait := 35 + rand.Intn(25)
			fmt.Printf("Czekam %d sekund przed następną księgą...\n", wait)
			time.Sleep(time.Duration(wait) * time.Second)
		}
	}
	fmt.Println("Koniec.")
}
