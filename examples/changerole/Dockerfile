FROM mcr.microsoft.com/dotnet/sdk:6.0-windowsservercore-ltsc2022

RUN curl https://download.microsoft.com/download/f/c/3/fc39707f-f67c-4c1b-9274-a055a3eb51b8/MicrosoftServiceFabric.7.2.477.9590.exe --output sf.exe

USER ContainerAdministrator
RUN sf.exe /quiet /accepteula /sdkcontainerclient
USER ContainerUser
RUN del /f /q sf.exe

ADD . C:/build
WORKDIR C:/build
RUN del /f /q changerole\PrintrolePkg\Code
RUN build.cmd

RUN mkdir C:\\app
RUN move changerole C:\\app
RUN del /f /q C:\build

WORKDIR C:/
ADD simplelauncher.exe C:/

CMD ["simplelauncher.exe", "app/changerole/PrintrolePkg"]




