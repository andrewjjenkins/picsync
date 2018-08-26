Picsync
=======

[![Docker Repository on
Quay](https://quay.io/repository/andrewjjenkins/picsync/status "Docker
Repository on Quay")](https://quay.io/repository/andrewjjenkins/picsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/andrewjjenkins/picsync)](https://goreportcard.com/report/github.com/andrewjjenkins/picsync)


This syncs pictures from a SmugMug album to a Nixplay Seed wifi-enabled picture
frame.

Setup
-----

You need a `picsync-config.yaml` file with credentials.

Start by running the OAuth setup for Smugmug.  Get a SmugMug API key from
https://api.smugmug.com/api/developer/apply and then perform a one-time OAuth
login:

```
NIXPLAY_SMUGMUG_API_KEY="fs..." NIXPLAY_SMUGMUG_API_SECRET="lfjd..." picsync login
```

Then, add the output of that command (a YAML block with smugmug credentials) to
a block with your nixplay credentials, in a file called `picsync-config.yaml`:

```yaml
# Keep this file confidential.
# If you lose it, de-authorize nixplay-sync from your SmugMug account and repeat 'picsync login'
smugmug:
  access:
    token: "V..."
    secret: "q..."
  consumer:
    token: "q..."
    secret: "n..."

# Just the username/password for your nixplay account
nixplay:
  username: "yourusername"
  password: "password"
```

Running
-------

Once you have setup, just run it, giving the name of a SmugMug Gallery to
sync to nixplay:

```
picsync sync "Smugmug Gallery" "Nixplay Album"
```

This will create a Nixplay album and sync all photos from `Smugmug Gallery` to
`Nixplay Album`, and then sync all those album photos to the slideshow
`ss_Nixplay Album`.

Go to https://app.nixplay.com/#/frames/ and click on a frame, then click "Enable
Playlist" for `ss_Nixplay Album` and it should automatically sync.

You can cause the syncer to run periodically:

```
picsync sync --every 60s "Smugmug Gallery" "Nixplay Album"
```

Roadmap/Help Wanted
-------------------
1. Delete pictures from nixplay when they're deleted from smugmug
2. Docker container or run in some FaaS somewhere (or let Smugmug/Nixplay run
   it)

License
-------
Copyright 2018 Andrew Jenkins <andrewjjenkins@gmail.com>

Licensed under the terms of the Apache 2.0 license (see LICENSE)