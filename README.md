🌩️ MyWeatherDash 🌩️

A modern Go-powered weather dashboard using WeeWX + MariaDB + Chart.js

MyWeatherDash is a personal weather dashboard powered by data from a WeeWX weather station.

The backend is written in Go, serving JSON data from a MariaDB (WeeWX) archive, and the frontend is a fully dynamic HTML/JS interface using Chart.js, day/night shading, wind vector overlays, and more.

This project is designed for:

Real-time and historical weather visualization

Custom dashboards beyond the default WeeWX skins

Lightweight, fast, and self-hosted deployments

Clear, modular Go API endpoints

✨ Features

📡 Live Data via Go API

Temperature & Dewpoint

Feels-like (Heat Index / Wind Chill)

Barometric pressure

Wind speed + gusts

Wind direction visualization

Rain rate & rain amount

Lightning strike totals

Inside temperature & humidity

Day/night chart shading

Master time grid alignment (all charts sync perfectly)

📈 Fully Dynamic Charts (Chart.js)

Shared time axis

Smoothed lines

Zero-value padding fixed

Rain and lightning graphs with intelligent scaling

Wind vector arrows

Wind direction scatter colored by speed

Automatic range selection (Day / Week / Month)

🖥 Current Conditions Panel

Displays:

Temperature

Dewpoint

Barometer

Active feels-like type

Wind speed + direction

Inside conditions

Rain totals (future feature)

🛠️ Project Structure

MyWeatherDash/

│

├── main.go

├── handlers.go

├── types.go

├── config.go

├── templates/

│   └── index.html

├── config.example.yaml

├── .gitignore

└── README.md

🔧 Setup

1. Clone the repo

git clone https://github.com/thurmanmarka/MyWeatherDash

cd MyWeatherDash

2. Create your real config file

Copy the example:

cp config.example.yaml config.yaml

Edit with your actual database credentials:

db:
  user: example_user
  
  password: "YOUR_PASSWORD"
  
  host: "127.0.0.1"
  
  port: 3306
  
  name: "weewx"
  
  params: "parseTime=false"


⚠️ config.yaml is ignored by Git to keep your credentials secure.

▶️ Run the Server

Use:

go run .


or build:

go build -o weatherdash
./weatherdash


The server starts on:

http://localhost:8080

🧪 API Endpoints

  All endpoints support:

  ?range=day
  
  ?range=week
  
  ?range=month

🏗 Future Improvements

  🗲 Modularized frontend (multiple JS files)

  🗲 Dark mode

  🗲 Systemd service for auto-start

  🗲 Docker container

  🗲 Configuration UI

  🗲 Alerts / notifications

📜 License

  This project is for personal use.

🙌 Contributing

  This is a personal project, but contributions or ideas are welcome through GitHub issues.
