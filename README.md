# 15 Puzzle game with Go

🚀 [Play in Telegram](https://t.me/puzzle_15_bot?startapp).
Game has [simple rules](https://en.wikipedia.org/wiki/15_puzzle).

## Description

The intentions of the project is to create a simple proof-of-concept demonstrating that a [Telegram Mini App](https://core.telegram.org/bots/webapps) can be developed with Go and [WebAssembly](https://go.dev/wiki/WebAssembly).

✨ Graphics are made with [Ebitengine](https://github.com/hajimehoshi/ebiten).

Basic "features" include: a splash screen, game move sound, a switchable silent mode without moves count,
a players' rating table, a congratulations screen for achieving 1st place,
a pin-code protected game statistics screen,
and backend API requests secured through [data validation](https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app).

## Running game

Local gameplay can be initiated using the following command:

```shell
go run ./cmd/ui
```

A fully functional Mini App requires a web server.
You can use [`Dockerfile`](Dockerfile) to build a Docker Image
containing all necessary artifacts: the server binary, the WebAssemply (wasm) binary, and html page.
Certain environment variables should be set for the server to start:

| Env                | Description |
| -                  | -           |
| **`BOT_TOKEN`**    | Telegram Bot [API Token](https://core.telegram.org/bots/tutorial). |
| **`DATA_FILE`**    | Path to the file where games data will be stored. |
| **`ACCESS_CODE`**  | Pin-code to access the Statistics Screen. |
| **`PROJECT_LINK`** | URL to the project's source code. |
| `SERVER_PORT`      | Port for the server to listen for API requests, defaulting to `8080` if not set. |
| `CONTEXT_ROOT`     | Server requests URI root path, defaulting to `/` if not set. |
| `STATIC_DIR`       | Directory where static files are located, defaulting to the current directory if not set. |

## Credits

- [Ebitengine](https://github.com/hajimehoshi/ebiten) game engine by Hajime Hoshi.
- Font by [VileR](http://int10h.org) and [fly_indiz](http://old-dos.ru/index.php?page=files&mode=files&do=show&id=102798).
- ASCII Gopher by [gheimifurt](https://www.reddit.com/r/golang/comments/xdxb9a/gopher_ascii_art_for_bashrc/).
