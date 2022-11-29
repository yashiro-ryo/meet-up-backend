# サーバー起動
run:
	go fmt . && go run .

dev:
	export BuildEnv=dev && make run

prod:
	export BuildEnv=prod && make run