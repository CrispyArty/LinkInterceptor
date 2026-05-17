; ---------------------------------------------------------
; NSIS Installer Script for Link Interceptor
; ---------------------------------------------------------

!include "MUI2.nsh"

; General Settings
Name "Link Interceptor"
OutFile "LinkInterceptorSetup.exe"
InstallDir "$PROGRAMFILES64\LinkInterceptor"
RequestExecutionLevel admin ; Required for HKLM registry access

; UI Settings
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Install"
    ; 1. Copy Files
    SetOutPath "$INSTDIR"
    ; Replace "dist/link_interceptor.exe" with the actual path to your Go binary
    File "link_interceptor.exe" 

    ; 2. Register as a URL Protocol Handler (ProgID)
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor" "" "URL:link-interceptor"
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor" "FriendlyTypeName" "Link Interceptor"
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor" "URL Protocol" ""
    
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor\Application" "ApplicationName" "Link Interceptor"
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor\DefaultIcon" "" "$INSTDIR\link_interceptor.exe,0"
    
    WriteRegStr HKLM "SOFTWARE\Classes\link-interceptor\shell\open\command" "" '"$INSTDIR\link_interceptor.exe" "%1"'

    ; 3. Register as a Browser (StartMenuInternet)
    ; This allows Windows to see your app in the "Default Browser" list
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities" "ApplicationDescription" "Link Interceptor"
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities" "ApplicationIcon" "$INSTDIR\link_interceptor.exe,0"
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities" "ApplicationName" "Link Interceptor"

    ; URL Associations for Browser selection
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities\URLAssociations" "https" "link-interceptor"
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities\URLAssociations" "http" "link-interceptor"

    ; Shell command for the browser entry
    WriteRegStr HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\shell\open\command" "" "$INSTDIR\link_interceptor.exe"

    ; 4. Register in "RegisteredApplications"
    ; This makes the app show up in the Windows "Default Apps" settings page
    WriteRegStr HKLM "SOFTWARE\RegisteredApplications" "Link_Interceptor" "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor\Capabilities"

    ; Create uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
    ; Remove Files
    Delete "$INSTDIR\link_interceptor.exe"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir "$INSTDIR"

    ; Remove Registry Keys (Clean up exactly what we added)
    DeleteRegKey HKLM "SOFTWARE\Classes\link-interceptor"
    DeleteRegKey HKLM "SOFTWARE\Clients\StartMenuInternet\Link_Interceptor"
    DeleteRegValue HKLM "SOFTWARE\RegisteredApplications" "Link_Interceptor"
SectionEnd