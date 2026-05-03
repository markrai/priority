# Project Priority Stack

A small web app for ordering projects. Data is stored in a JSON file on the server.

## Run with Docker

Build the image from this folder:

```bash
docker build -t priority-stack .
```

Run it. The app listens on port **8080** inside the container; **8095** is the port on your machine or NAS. The folder after `-v` is where `projects.json` is kept (create it first if needed).

```bash
docker run -p 8095:8080 -v /volume1/docker/priority/data:/app/data priority-stack
```

On a Synology NAS, `/volume1/docker/priority/data` is a typical place for that data folder. Use any path you like as long as it exists (or Docker can create the mount point as required).

The container runs as **root** so a bind-mounted `data` folder on DSM is usually writable. If you override `user:` in Compose, `chown` that host folder to the same uid/gid or POSTs will fail with **500 write failed**.

Then open **http://\<your-device-ip\>:8095** in a browser.

Optional: use `docker-compose.yml` in this repo for the same idea with Compose or Portainer stacks.
