dotnet restore %~dp0\src\changerole\Printrole\Printrole.csproj -s https://api.nuget.org/v3/index.json
dotnet build %~dp0\src\changerole\Printrole\Printrole.csproj -v normal

for %%F in ("%~dp0\src\changerole\Printrole\Printrole.csproj") do cd %%~dpF
dotnet publish -o %~dp0\changerole\PrintrolePkg\Code
cd %~dp0