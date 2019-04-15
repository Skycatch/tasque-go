FROM microsoft/nanoserver

ARG NODE_VERSION=4.8.3
ADD https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-win-x64.zip C:/node.zip
ENV NODE_VERSION 4.8.3
RUN powershell -Command Expand-Archive C:/node.zip C:/
RUN powershell -Command Rename-Item C:/node-v$env:NODE_VERSION-win-x64 C:/node
RUN SETX PATH C:\node


#FROM microsoft/nanoserver
#
#SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
#
#ENV NPM_CONFIG_LOGLEVEL info
#
#COPY --from=download /nodejs /nodejs
#
#RUN New-Item $($env:APPDATA + '\npm') ; \
#    $env:PATH = 'C:\nodejs;{0}\npm;{1}' -f $env:APPDATA, $env:PATH ; \
#    Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\' -Name Path -Value $env:PATH
#
COPY ./tasque c:/tasque
ENTRYPOINT [ "c:/tasque" ]