#Requires -Version 5
param($install_dir)

# This script can be used to download and install src-fingerprint on Windows,
# from a PowerShell prompt. Refer to src-fingerprint README for more details.

$DEFAULT_INSTALL_DIR = "$env:APPDATA\src-fingerprint"

# Quit if anything goes wrong
$old_erroractionpreference = $erroractionpreference
$erroractionpreference = 'stop'

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run this installer."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

# Required to uncompress zip files
Add-Type -Assembly "System.IO.Compression.FileSystem"

function request_install_dir() {
    $dir = Read-Host -Prompt "Installation dir [$DEFAULT_INSTALL_DIR]"
    if ($dir -eq "") {
        $dir = $DEFAULT_INSTALL_DIR
    }
    return $dir
}

function recreate_dir($dir) {
    if (test-path $dir) {
        Remove-Item $dir -Recurse -Force
    }
    mkdir $dir > $null
}

function download($url, $dst_file) {
    $client = New-Object Net.Webclient
    $client.DownloadFile($url, $dst_file)
}

function extract($zip_file, $dst_dir) {
    [IO.Compression.ZipFile]::ExtractToDirectory($zip_file, $dst_dir)
}

function get_latest_version() {
    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/GitGuardian/src-fingerprint/releases/latest"
    $tag = $response.tag_name
    if ($tag.Length -lt 2 -or $tag[0] -ne "v") {
        Write-Error -Category InvalidResult -Message "Invalid tag name: '$tag'"
        exit
    }
    # $tag is "v1.2.3", we want "1.2.3"
    $tag.SubString(1)
}

function main() {
    if ($install_dir -eq $null) {
        $install_dir = request_install_dir
    }

    $tmp_dir = "$install_dir\tmp"
    recreate_dir $tmp_dir

    Write-Output "Fetching version number of latest release"
    $version = get_latest_version
    Write-Output "Latest version is $version"

    $zip_url = "https://github.com/GitGuardian/src-fingerprint/releases/download/v${version}/src-fingerprint_${version}_Windows_amd64.zip"

    Write-Output "Downloading $zip_url"
    $zip_file = "$tmp_dir\install.zip"
    download $ZIP_URL $zip_file

    Write-Output 'Extracting'
    extract $zip_file $tmp_dir
    Move-Item -Path "$tmp_dir\src-fingerprint_${VERSION}_Windows_*\*" -Destination $install_dir -Force

    Write-Output 'Cleaning'
    Remove-Item $tmp_dir -Recurse -Force

    Write-Output "All done, src-fingerprint is now available at $install_dir\src-fingerprint.exe"
}

main

# Reset $erroractionpreference to original value
$erroractionpreference = $old_erroractionpreference
