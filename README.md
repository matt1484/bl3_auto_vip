# BL3 Auto VIP

Cross platform .NET Core app for automatically redeeming VIP codes for Borderlands 3

Current zips of standalone executable can be found at
https://drive.google.com/drive/folders/1cFpaUXDhXYTHUSFXvQPV1NCK7-E5N619

Requires .NET core 3.0 (or you can change the version in the .csproj)

To run from source:

1. download project
2. navigate to project
3. run `dotnet restore`
4. run `dotnet run`


Note: the first time running make take longer than expected since it needs to download a few extra things,

Note: this program creates a folder named `bl3-auto-vip-browsers` at `Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);`
for Mac/Windows and `$HOME/.local/share/` for linux. This should be deleted if you no longer wish to use this program.
