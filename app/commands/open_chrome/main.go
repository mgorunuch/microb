package main

import (
	"context"
	"flag"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jackc/pgx/v5"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/postgres"
)

var controlHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Review Control Panel</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h2 {
            margin: 0 0 20px 0;
            color: #333;
        }
        .info-box {
            margin-bottom: 20px;
            padding: 15px;
            background-color: #e3f2fd;
            border-radius: 4px;
            border-left: 4px solid #1976d2;
        }
        .url-display {
            margin-bottom: 20px;
            padding: 10px;
            background-color: #f8f9fa;
            border-radius: 4px;
            word-break: break-all;
        }
        .window-info {
            margin-bottom: 20px;
            font-size: 14px;
            color: #666;
        }
        .btn {
            padding: 12px 24px;
            background-color: #4CAF50;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            transition: all 0.3s ease;
        }
        .btn:hover {
            background-color: #45a049;
            transform: translateY(-1px);
        }
        .btn:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        .progress {
            margin-bottom: 20px;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Review Control Panel</h2>
        <div class="info-box">
            <strong>Window 1:</strong> Chrome browser with the target page<br>
            <strong>Window 2 (this):</strong> Control panel
        </div>
        <div class="progress">
            URL {{.Current}} of {{.Total}}
        </div>
        <div class="url-display">
            <strong>Currently reviewing:</strong><br>
            {{.URL}}
        </div>
        <button class="btn" onclick="finishReview()" id="doneBtn">Done Reviewing</button>
    </div>
    <script>
        function finishReview() {
            const btn = document.getElementById('doneBtn');
            btn.disabled = true;
            btn.textContent = 'Saving...';
            
            fetch('/done', {
                method: 'POST'
            }).then(() => {
                window.close();
            });
        }
    </script>
</body>
</html>
`

type ControlServer struct {
	doneChan chan bool
	url      string
	current  int
	total    int
	tmpl     *template.Template
}

func NewControlServer() (*ControlServer, error) {
	tmpl, err := template.New("control").Parse(controlHTML)
	if err != nil {
		return nil, err
	}

	return &ControlServer{
		doneChan: make(chan bool),
		tmpl:     tmpl,
	}, nil
}

func (s *ControlServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/done" {
		s.doneChan <- true
		w.WriteHeader(http.StatusOK)
		return
	}

	s.tmpl.Execute(w, struct {
		URL     string
		Current int
		Total   int
	}{
		URL:     s.url,
		Current: s.current,
		Total:   s.total,
	})
}

func main() {
	reason := flag.String("reason", "manual check", "Reason for visiting URLs")
	flag.Parse()

	core.Init()
	cleanup := postgres.Init(context.Background())
	defer cleanup()

	// Create Chrome instance
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.WindowSize(1200, 800),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Create control server
	server, err := NewControlServer()
	if err != nil {
		core.Logger.Fatal(err)
	}

	// Count total URLs
	var urls []string
	countChan := make(chan string)
	go func() {
		for url := range countChan {
			urls = append(urls, url)
		}
		server.total = len(urls)
	}()
	core.ProcessLines(core.SimpleConfig[bool]{
		ThreadsCount: 1,
		RunFunc: func(input string) (bool, error) {
			countChan <- input
			return true, nil
		},
	})
	close(countChan)

	// Start HTTP server
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe("localhost:5322", server); err != nil {
			core.Logger.Fatal(err)
		}
	}()

	// Process URLs one by one
	for i, input := range urls {
		server.current = i + 1
		url := strings.TrimSpace(input)

		// Check if URL was visited in the last month
		lastVisit, err := postgres.GetLastVisit(ctx, url)
		if err != nil && err != pgx.ErrNoRows {
			core.Logger.Errorf("Failed to check last visit: %v", err)
			continue
		}
		if lastVisit != nil {
			monthAgo := time.Now().AddDate(0, -1, 0)
			if lastVisit.OpenedAt.After(monthAgo) {
				core.Logger.Infof("Skipping URL %s - already visited on %s", url, lastVisit.OpenedAt)
				continue
			}
		}

		visit := &postgres.ChromeVisit{
			URL:      &postgres.URL{Raw: url},
			URLs:     []*postgres.URL{},
			OpenedAt: time.Now(),
			Success:  true,
			Reason:   *reason,
		}

		// Create new tab context
		tabCtx, cancel := chromedp.NewContext(ctx)

		// Add timeout - 5 minutes for review
		tabCtx, cancel = context.WithTimeout(tabCtx, 5*time.Minute)

		core.Logger.Infof("Opening URL %d of %d: %s", i+1, len(urls), url)

		// Navigate to the target URL in main tab
		err = chromedp.Run(tabCtx,
			chromedp.Navigate(url),
			chromedp.Title(&visit.Title),
		)

		if err != nil {
			visit.Success = false
			visit.ErrorMsg = err.Error()
			core.Logger.Errorf("Failed to process URL %s: %v", url, err)
			cancel()
			continue
		}

		// Update control server URL
		server.url = url

		// Open control page in default browser
		core.Logger.Info("Opening control panel...")
		if err := chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:5322"),
		); err != nil {
			core.Logger.Error("Failed to open control panel:", err)
		}

		// Wait for done signal
		<-server.doneChan

		// Capture HTML after review
		err = chromedp.Run(tabCtx,
			chromedp.OuterHTML("html", &visit.HTML),
		)

		if err != nil {
			visit.Success = false
			visit.ErrorMsg = err.Error()
			core.Logger.Errorf("Failed to capture HTML: %v", err)
		}

		// Save to database using the existing postgres package
		if err := postgres.CreateChromeVisit(ctx, visit); err != nil {
			core.Logger.Errorf("Failed to save to database: %v", err)
		}

		core.Logger.Infof("Processed URL: %s - Title: %s", url, visit.Title)
		cancel()
	}

	core.Logger.Info("All URLs processed")
	core.Logger.Info("Closing control panel...")
}
