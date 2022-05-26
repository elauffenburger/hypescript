$ErrorActionPreference = "Stop"

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

try {
    # Navigate back to the root dir.
    pushd (Join-Path $PSScriptRoot "../")

    # Run the compiler.
    go run .
    if ($LASTEXITCODE -ne 0) {
        throw "Compilation failed!"
    }

    # Set up the session as a Visual Studio dev environment.
    Write-VsVars

    pushd "./build"

    try {
        $Srcs = Get-ChildItem "." -Filter "*.cpp" | % { $_.FullName }
        cl $Srcs /link /DEBUG:FULL /out:main.exe
        if ($LASTEXITCODE -ne 0) {
            throw "Compilation failed!"
        }

        ./main.exe
    }
    finally {
        popd
    }
}
finally {
    popd
}
