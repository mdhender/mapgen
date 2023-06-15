# mapgen
A fantasy map generator.

# Running

## Mac or Linux
1. Clone the repository.
2. Open a terminal and navigate to the root of the repository.
3. Build the executable with `go build`.
4. Change to the `testdata` directory with `cd testdata`.
5. Start the `mapgen` executable by running `../mapgen`.

## Windows
1. Clone the repository.
2. Open a terminal and navigate to the root of the repository.
3. Build the executable with `go build`.
4. Change to the `testdata` directory with `cd testdata`.
5. Start the `mapgen` executable by running `..\mapgen.exe`.

# Viewing
Open `http://localhost:8080/` in your browser.

## First time
Note that it takes about ten seconds to generate a new image for any seed:

    2023/06/14 17:32:18 POST /generate: json: created c0ffeecafe.json
    2023/06/14 17:32:18 POST /generate: elapsed 6.342958125sn

The results are cached so that viewing or customizing for the same seed value happens in a fraction of a second:

    2023/06/14 17:32:39 POST /generate: entering
    2023/06/14 17:32:39 POST /generate: elapsed 251.442791msn

# Setting the Secret
To set the secret,
set and export the environment variable `MAPGEN_SECRET`,
then start the server.