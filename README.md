# BL3 Auto VIP + Shift

Cross platform Go app for automatically redeeming VIP/Shift codes
for Borderlands 3. Also redeems VIP activities such as weekly twitter/facebook points.

## Getting Started

1. Make a VIP account at https://borderlands.com/en-US/vip/
2. Download program from above link
3. Unzip the folder
4. Run it, you will be prompted for username and password
5. Enter username and password (we only use this info to sign into borderlands)
6. Watch it do its magic
7. Repeat when more codes come out


Run it with `--help` to view command line args that are supported.

### Installing

#### Using go
```sh
go get -u github.com/matt1484/bl3_auto_vip
```

#### Docker
To run from source:
1. install docker
2. download project
3. navigate to project
4. run `docker build -t bl3 .`
5. run `docker run -it bl3`

#### Using the prebuilt releases
The binaries/executables are released
[here](https://github.com/matt1484/bl3_auto_vip/releases)

## FAQs

### Why does my operating system say it's an unrecgonized app?
Telling the operating system that we're a trusted source is expensive.
This is a small open source project and we don't have the funds to correctly
sign the app.

### Why does my antivirus flag this program?
It's a false positive. If you don't trust us, you can look at the code and
compile it yourself. That's one of the beauties of an open source project!

### It's not working. What should I do?
File an issue here with as much detail as you can provide. We're working on
adding additional logging and a bug template to better assist with any issues.

## License
This project is licensed under the Apache-2.0 License - see the
[LICENSE](LICENSE) file for details
