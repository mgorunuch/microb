package main

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/chrome"
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
            URL {{.Current}}
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
	tmpl     *template.Template
	ctx      *chrome.Context
}

func NewControlServer(ctx *chrome.Context) (*ControlServer, error) {
	tmpl, err := template.New("control").Parse(controlHTML)
	if err != nil {
		return nil, err
	}

	return &ControlServer{
		doneChan: make(chan bool),
		current:  1,
		tmpl:     tmpl,
		ctx:      ctx,
	}, nil
}

func (s *ControlServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/done" {
		s.Done()
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
	})
}
func (s *ControlServer) GetHostname() string {
	return `localhost:5322`
}

func (s *ControlServer) Done() {
	s.doneChan <- true
}

func (s *ControlServer) Wait() {
	<-s.doneChan
}

func (s *ControlServer) ListenAndServe(wg *sync.WaitGroup) func() {
	wg.Add(1)

	return func() {
		defer wg.Done()
		if err := http.ListenAndServe(s.GetHostname(), s); err != nil {
			core.Logger.Fatal(err)
		}
	}
}

func (s *ControlServer) SetURL(url string) {
	s.url = url
}

func (s *ControlServer) NavigateToURL() error {
	return s.ctx.Run(chromedp.Navigate(s.url))
}

func (s *ControlServer) ApplyCurrentUrl(url string) error {
	s.SetURL(url)
	return s.NavigateToURL()
}

func (s *ControlServer) IncrCurrent() {
	s.current += 1
}
