# BL3 Auto VIP

Cross platform Go app for automatically redeeming VIP codes for Borderlands 3

Current zips of standalone executable can be found at
http://www.mediafire.com/folder/aatm7o7bc9eij/bl3-auto-vip


To run from source:
1. install docker
2. download project
3. navigate to project
4. run `docker build -t bl3 .`
5. run `docker run -it bl3`

Run it with `--help` to view command line args that are supported.

Update Log: 

* v1.0: Initial release 
* v1.1: Fixed timeout issues and added support for command line args (email and password) p.s. it is also much faster
* v1.2: Added a timer so it does not immediately close when done and also added support for codes with multiple types
* v1.2.1: Fixed bug where tables in comments would count as codes and add password masking
* v1.3: Rewrote all code in go to add future mobile support (also more maintainable and smaller executable)

To do:
* shift codes redemption
* auto facebook/instagram/twitter weekly points (cant watch videos but can read the articles maybe)
* fake emails for referral points (captcha is making this hard, so maybe not...)
* look into other login types (PSN, Xbox, etc.)
* android/ios version 
