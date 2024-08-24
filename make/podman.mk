podman/up:
	podman-compose up -d
	podman ps

podman/resource:
	podman-compose up -d redis mariadb postgres influx

podman/down:
	podman-compose down

podman/prune:
	podman system prune --a --volumes
