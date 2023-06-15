# mapgen
A fantasy map generator.

# Building
Mapgen is written in Go.
You must install the Go compiler to build.

1. Clone the repository.
2. Open a terminal and navigate to the root of the repository.
3. Build the executable with `go build`.

# Running
NOTE:
The example command lines use "water.slide" as the secret.
Please use your own secret!

## Mac or Linux
1. Open a terminal and navigate to the root of the repository.
2. Change to the `testdata` directory with `cd testdata`.
3. Start the `mapgen` executable by running `../mapgen --secret water.slide`.

The program will display the secret at startup:

    2023/06/15 13:50:24 mapgen: secret "water.slide"
    2023/06/15 13:50:24 static: file: ../public/favicon.ico

## Windows
1. Open a terminal and navigate to the root of the repository.
2. Change to the `testdata` directory with `cd testdata`.
3. Start the `mapgen` executable by running `..\mapgen.exe --secret water.slide`.

The program will display the secret at startup:

    2023/06/15 13:50:24 mapgen: secret "water.slide"
    2023/06/15 13:50:24 static: file: ../public/favicon.ico

# Viewing
Open http://localhost:8080/ in your browser.

## First time
Note that it can take up to twenty seconds to generate a new image for any seed:

    2023/06/14 17:32:18 POST /generate: json: created 783762.json
    2023/06/14 17:32:18 POST /generate: elapsed 16.342958125sn

The results are cached so that viewing or customizing for the same seed value happens in a fraction of a second:

    2023/06/14 17:32:39 POST /generate: entering
    2023/06/14 17:32:39 POST /generate: elapsed 251.442791msn
