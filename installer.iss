[Setup]
AppName=Golt Runtime
AppVersion=1.0.3
AppPublisher=Aztekode
AppPublisherURL=https://github.com/Aztekode/golt
AppId={{A4F7B7A7-9E0D-4C0D-9DA7-1C7C4C8C2F35}
DefaultDirName={autopf}\Golt
DefaultGroupName=Golt
DisableProgramGroupPage=yes
OutputBaseFilename=GoltSetup_v1.0.3_windows_amd64
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
ChangesEnvironment=yes
UninstallDisplayIcon={app}\golt.exe
SetupIconFile=assets\installer\golt.ico
WizardImageFile=assets\installer\wizard.bmp
WizardSmallImageFile=assets\installer\wizard-small.bmp

#define GoltPfxPath GetEnv("GOLT_PFX_PATH")
#define GoltPfxPass GetEnv("GOLT_PFX_PASSWORD")
#define GoltTimestampUrl GetEnv("GOLT_TIMESTAMP_URL")
#define GoltSigningEnabled (GoltPfxPath != "") && (GoltPfxPass != "")

#if GoltSigningEnabled
SignTool=signtool
SignedUninstaller=yes
#endif

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[Tasks]
Name: "envPath"; Description: "Add Golt to the PATH (Recommended)"; GroupDescription: "System Integration:"
Name: "examples"; Description: "Install examples"; GroupDescription: "Extras:"
Name: "desktopicon"; Description: "Create a desktop icon"; GroupDescription: "Shortcuts:"

[Files]
#if GoltSigningEnabled
Source: "golt.exe"; DestDir: "{app}"; Flags: ignoreversion; SignTool: signtool
#else
Source: "golt.exe"; DestDir: "{app}"; Flags: ignoreversion
#endif
Source: "examples\*"; DestDir: "{app}\examples"; Flags: ignoreversion recursesubdirs; Tasks: examples

[Icons]
Name: "{group}\Golt Runtime"; Filename: "{app}\golt.exe"
Name: "{group}\Uninstall Golt Runtime"; Filename: "{uninstallexe}"
Name: "{commondesktop}\Golt Runtime"; Filename: "{app}\golt.exe"; Tasks: desktopicon

[Code]
function GetPathRootKey(): Integer;
begin
  if IsAdminInstallMode() then
    Result := HKEY_LOCAL_MACHINE
  else
    Result := HKEY_CURRENT_USER;
end;

function GetPathSubkey(): string;
begin
  if IsAdminInstallMode() then
    Result := 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment'
  else
    Result := 'Environment';
end;

function GetAppRegSubkey(): string;
begin
  Result := 'Software\Aztekode\Golt';
end;

procedure WriteAddedToPathFlag();
begin
  RegWriteStringValue(GetPathRootKey(), GetAppRegSubkey(), 'AddedToPath', '1');
end;

function WasAddedToPath(): Boolean;
var
  Value: string;
begin
  if RegQueryStringValue(GetPathRootKey(), GetAppRegSubkey(), 'AddedToPath', Value) then
    Result := Value = '1'
  else
    Result := False;
end;

procedure AppendString(var Arr: TArrayOfString; Value: string);
var
  L: Integer;
begin
  L := GetArrayLength(Arr);
  SetArrayLength(Arr, L + 1);
  Arr[L] := Value;
end;

function NormalizePathList(Value: string): string;
var
  Parts: TArrayOfString;
  ResultParts: TArrayOfString;
  I: Integer;
  Part: string;
begin
  Parts := SplitString(Value, ';');
  SetArrayLength(ResultParts, 0);
  for I := 0 to GetArrayLength(Parts) - 1 do
  begin
    Part := Trim(Parts[I]);
    if Part = '' then
      continue;
    AppendString(ResultParts, Part);
  end;
  Result := JoinString(ResultParts, ';');
end;

function ReadPathValue(): string;
var
  OrigPath: string;
begin
  if RegQueryStringValue(GetPathRootKey(), GetPathSubkey(), 'Path', OrigPath) then
    Result := OrigPath
  else
    Result := '';
end;

procedure WritePathValue(Value: string);
begin
  RegWriteExpandStringValue(GetPathRootKey(), GetPathSubkey(), 'Path', Value);
end;

function HasPathEntry(PathValue: string; Entry: string): Boolean;
begin
  Result := Pos(';' + Lowercase(Entry) + ';', ';' + Lowercase(PathValue) + ';') > 0;
end;

procedure AddToPath(Entry: string);
var
  OrigPath: string;
  NewPath: string;
begin
  OrigPath := NormalizePathList(ReadPathValue());
  if HasPathEntry(OrigPath, Entry) then
    exit;

  if OrigPath = '' then
    NewPath := Entry
  else
    NewPath := OrigPath + ';' + Entry;

  WritePathValue(NewPath);
end;

procedure RemoveFromPath(Entry: string);
var
  OrigPath: string;
  Parts: TArrayOfString;
  ResultParts: TArrayOfString;
  I: Integer;
  Part: string;
begin
  OrigPath := NormalizePathList(ReadPathValue());
  Parts := SplitString(OrigPath, ';');
  SetArrayLength(ResultParts, 0);

  for I := 0 to GetArrayLength(Parts) - 1 do
  begin
    Part := Trim(Parts[I]);
    if Part = '' then
      continue;
    if Lowercase(Part) = Lowercase(Entry) then
      continue;
    AppendString(ResultParts, Part);
  end;

  WritePathValue(JoinString(ResultParts, ';'));
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if WizardIsTaskSelected('envPath') then
    begin
      AddToPath(ExpandConstant('{app}'));
      WriteAddedToPathFlag();
    end;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
  begin
    if WasAddedToPath() then
      RemoveFromPath(ExpandConstant('{app}'));
  end;
end;

[SignTool]
#if GoltSigningEnabled
#if GoltTimestampUrl != ""
Name: "signtool"; Parameters: "sign /fd sha256 /td sha256 /f ""{#GoltPfxPath}"" /p ""{#GoltPfxPass}"" /tr ""{#GoltTimestampUrl}"" ""$f"""
#else
Name: "signtool"; Parameters: "sign /fd sha256 /td sha256 /f ""{#GoltPfxPath}"" /p ""{#GoltPfxPass}"" ""$f"""
#endif
#endif
