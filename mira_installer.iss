[Setup]
AppName=Mira Programming Language
AppVersion=0.1
DefaultDirName={commonpf}\Mira
DefaultGroupName=Mira Programming Language
UninstallDisplayIcon={app}\mira.exe
OutputDir=.
OutputBaseFilename=MiraSetup
Compression=lzma
SolidCompression=yes

[Files]
Source: "mira.exe"; DestDir: "{commonpf}\Mira"; Flags: ignoreversion

[Icons]
Name: "{group}\Mira Programming Language"; Filename: "{commonpf}\Mira\mira.exe"

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{commonpf}\Mira"; Check: IsAdminInstallMode