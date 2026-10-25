; Inno Setup 6 script for the Cardflip installer. tool/package.sh compiles it
; with ISCC and passes everything that varies with /D:
;     AppVersion          the version, X.Y.Z
;     FileVersion         X.Y.Z.<build number>, for the setup's file version
;     Arch                x64 or arm64
;     SourceDir           the staged app folder, the same files as the zip
;     OutputDir           where the setup executable is written
;     OutputBaseFilename  its name without .exe

#ifndef AppVersion
  #error AppVersion is not defined; build the installer with tool/package.sh windows-installer
#endif
#ifndef FileVersion
  #error FileVersion is not defined
#endif
#ifndef Arch
  #error Arch is not defined
#endif
#ifndef SourceDir
  #error SourceDir is not defined
#endif
#ifndef OutputDir
  #error OutputDir is not defined
#endif
#ifndef OutputBaseFilename
  #error OutputBaseFilename is not defined
#endif

#if Arch == "arm64"
  #define ArchitectureId "arm64"
#elif Arch == "x64"
  #define ArchitectureId "x64compatible"
#else
  #error Arch must be x64 or arm64
#endif

[Setup]
; Identifies Cardflip to Windows: a newer installer with the same AppId
; upgrades the earlier install and keeps one entry in "Installed apps".
; It must never change.
AppId={{2926F336-6B2E-4A06-B1A1-32B8172D3B14}
AppName=Cardflip
AppVersion={#AppVersion}
AppVerName=Cardflip {#AppVersion}
AppPublisher=Cardflip
AppPublisherURL=https://github.com/madzione/cardflip
AppSupportURL=https://github.com/madzione/cardflip
VersionInfoVersion={#FileVersion}
; Per user in %LOCALAPPDATA%\Programs\Cardflip without elevation, or for all
; users in Program Files when chosen in the dialog or with /ALLUSERS.
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog commandline
DefaultDirName={autopf}\Cardflip
DisableProgramGroupPage=yes
ArchitecturesAllowed={#ArchitectureId}
ArchitecturesInstallIn64BitMode={#ArchitectureId}
; The Restart Manager asks to close a running Cardflip before files are
; replaced.
CloseApplications=yes
RestartApplications=no
SetupIconFile=..\runner\resources\app_icon.ico
UninstallDisplayIcon={app}\cardflip.exe
UninstallDisplayName=Cardflip
WizardStyle=modern
Compression=lzma2
SolidCompression=yes
OutputDir={#OutputDir}
OutputBaseFilename={#OutputBaseFilename}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "german"; MessagesFile: "compiler:Languages\German.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

; An upgrade replaces the assets, so none of an earlier version's stay. The
; collection is in %APPDATA%, outside {app}, and is never touched.
[InstallDelete]
Type: filesandordirs; Name: "{app}\data"

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\Cardflip"; Filename: "{app}\cardflip.exe"
Name: "{autodesktop}\Cardflip"; Filename: "{app}\cardflip.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\cardflip.exe"; Description: "{cm:LaunchProgram,Cardflip}"; Flags: nowait postinstall skipifsilent
