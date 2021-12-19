function Write-VsVars() {
    pushd "C:\Program Files (x86)\Microsoft Visual Studio\2017\Community\Common7\Tools"
    cmd /c "VsDevCmd.bat&set" |
    foreach {
        if ($_ -match "=") {
            $v = $_.split("="); set-item -force -path "ENV:\$($v[0])"  -value "$($v[1])"
        }
    }
    popd

    Write-Host "`nVisual Studio 2017 Command Prompt variables set." -ForegroundColor Yellow
}

go run .

Write-VsVars

$Srcs = Get-ChildItem "./build" -Filter "*.cpp" | %{ $_.FullName }

cl $Srcs /link /out:./build/main.exe

./build/main.exe
