Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Workspace = Split-Path -Parent $ScriptDir
$OutDir = $ScriptDir
$DeckPath = Join-Path $OutDir "MCA_Final_Project_Presentation_Parag_Gulati.pptx"
$ReadmePath = Join-Path $OutDir "PRESENTATION_README.md"
$StageDir = Join-Path $Workspace ".tmp\mca_final_project_pptx_stage"

$SlideW = 13.333
$SlideH = 7.5
$EmuPerIn = 914400

$Colors = @{
  Ink = "132033"
  Muted = "64748B"
  Navy = "12355B"
  Blue = "2563EB"
  Teal = "0F766E"
  Green = "16A34A"
  Amber = "F59E0B"
  Red = "DC2626"
  Violet = "6D28D9"
  Sky = "0284C7"
  Light = "F8FAFC"
  Panel = "FFFFFF"
  Line = "CBD5E1"
  SoftBlue = "EFF6FF"
  SoftTeal = "ECFDF5"
  SoftAmber = "FFFBEB"
  SoftRed = "FEF2F2"
  SoftViolet = "F5F3FF"
  Dark = "0F172A"
}

function Assert-InWorkspace {
  param([string]$Path)
  $full = [System.IO.Path]::GetFullPath($Path)
  $root = [System.IO.Path]::GetFullPath($Workspace)
  if (-not $full.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to operate outside workspace: $full"
  }
}

function Reset-Stage {
  Assert-InWorkspace $StageDir
  if (Test-Path $StageDir) {
    Remove-Item -LiteralPath $StageDir -Recurse -Force
  }
  New-Item -ItemType Directory -Force -Path $StageDir | Out-Null
}

function Write-Utf8 {
  param([string]$Path, [string]$Content)
  $dir = Split-Path -Parent $Path
  New-Item -ItemType Directory -Force -Path $dir | Out-Null
  [System.IO.File]::WriteAllText($Path, $Content, [System.Text.UTF8Encoding]::new($false))
}

function XmlEsc {
  param([AllowNull()][string]$Value)
  if ($null -eq $Value) { return "" }
  return [System.Security.SecurityElement]::Escape($Value)
}

function Emu {
  param([double]$Inches)
  return [int64][Math]::Round($Inches * $EmuPerIn)
}

function PxSz {
  param([int]$Pt)
  return $Pt * 100
}

function FillXml {
  param([string]$Color)
  if ([string]::IsNullOrWhiteSpace($Color) -or $Color -eq "none") {
    return "<a:noFill/>"
  }
  return "<a:solidFill><a:srgbClr val=""$Color""/></a:solidFill>"
}

function LineXml {
  param(
    [string]$Color = "none",
    [int]$Width = 1,
    [bool]$Dash = $false
  )
  $w = [Math]::Max(1, $Width) * 12700
  if ([string]::IsNullOrWhiteSpace($Color) -or $Color -eq "none") {
    return "<a:ln w=""$w""><a:noFill/></a:ln>"
  }
  $dashXml = ""
  if ($Dash) { $dashXml = "<a:prstDash val=""dash""/>" }
  return "<a:ln w=""$w""><a:solidFill><a:srgbClr val=""$Color""/></a:solidFill>$dashXml</a:ln>"
}

function RunXml {
  param(
    [string]$Text,
    [int]$Size = 18,
    [string]$Color = "132033",
    [bool]$Bold = $false
  )
  $b = ""
  if ($Bold) { $b = " b=""1""" }
  $sz = PxSz $Size
  $safe = XmlEsc $Text
  return "<a:r><a:rPr lang=""en-US"" sz=""$sz"" dirty=""0""$b><a:solidFill><a:srgbClr val=""$Color""/></a:solidFill><a:latin typeface=""Aptos""/><a:cs typeface=""Aptos""/></a:rPr><a:t>$safe</a:t></a:r>"
}

function ParagraphXml {
  param(
    [string]$Text,
    [int]$Size = 18,
    [string]$Color = "132033",
    [bool]$Bold = $false,
    [string]$Align = "l"
  )
  return "<a:p><a:pPr algn=""$Align""/>$(RunXml -Text $Text -Size $Size -Color $Color -Bold $Bold)</a:p>"
}

function TextBodyXml {
  param(
    [string[]]$Lines,
    [int]$Size = 18,
    [string]$Color = "132033",
    [bool]$Bold = $false,
    [string]$Align = "l",
    [string]$Anchor = "t",
    [int]$Inset = 8
  )
  $in = [int]($Inset * 12700)
  $paras = New-Object System.Collections.Generic.List[string]
  foreach ($line in $Lines) {
    $paras.Add((ParagraphXml -Text $line -Size $Size -Color $Color -Bold $Bold -Align $Align))
  }
  if ($paras.Count -eq 0) {
    $paras.Add("<a:p/>")
  }
  return "<p:txBody><a:bodyPr wrap=""square"" lIns=""$in"" rIns=""$in"" tIns=""$in"" bIns=""$in"" anchor=""$Anchor""><a:normAutofit fontScale=""88000"" lnSpcReduction=""12000""/></a:bodyPr><a:lstStyle/>$($paras -join '')</p:txBody>"
}

function NextShapeId {
  $script:ShapeId += 1
  return $script:ShapeId
}

function Add-Shape {
  param(
    [double]$X,
    [double]$Y,
    [double]$W,
    [double]$H,
    [string[]]$Text = @(),
    [string]$Fill = "FFFFFF",
    [string]$Line = "CBD5E1",
    [int]$LineWidth = 1,
    [string]$Geom = "roundRect",
    [int]$FontSize = 16,
    [string]$FontColor = "132033",
    [bool]$Bold = $false,
    [string]$Align = "l",
    [string]$Anchor = "mid",
    [int]$Inset = 8,
    [bool]$Dash = $false
  )
  $id = NextShapeId
  $x = Emu $X; $y = Emu $Y; $w = Emu $W; $h = Emu $H
  $tx = ""
  if ($Text.Count -gt 0) {
    $tx = TextBodyXml -Lines $Text -Size $FontSize -Color $FontColor -Bold $Bold -Align $Align -Anchor $Anchor -Inset $Inset
  }
  $xml = @"
<p:sp>
  <p:nvSpPr><p:cNvPr id="$id" name="Shape $id"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
  <p:spPr><a:xfrm><a:off x="$x" y="$y"/><a:ext cx="$w" cy="$h"/></a:xfrm><a:prstGeom prst="$Geom"><a:avLst/></a:prstGeom>$(FillXml $Fill)$(LineXml -Color $Line -Width $LineWidth -Dash $Dash)</p:spPr>
  $tx
</p:sp>
"@
  $script:Shapes.Add($xml) | Out-Null
}

function Add-Text {
  param(
    [double]$X,
    [double]$Y,
    [double]$W,
    [double]$H,
    [string[]]$Lines,
    [int]$FontSize = 18,
    [string]$Color = "132033",
    [bool]$Bold = $false,
    [string]$Align = "l",
    [string]$Anchor = "t",
    [int]$Inset = 3
  )
  Add-Shape -X $X -Y $Y -W $W -H $H -Text $Lines -Fill "none" -Line "none" -Geom "rect" -FontSize $FontSize -FontColor $Color -Bold $Bold -Align $Align -Anchor $Anchor -Inset $Inset
}

function Add-Line {
  param(
    [double]$X1,
    [double]$Y1,
    [double]$X2,
    [double]$Y2,
    [string]$Color = "94A3B8",
    [int]$Width = 2,
    [bool]$Arrow = $true
  )
  $id = NextShapeId
  $x = Emu ([Math]::Min($X1, $X2))
  $y = Emu ([Math]::Min($Y1, $Y2))
  $cx = [Math]::Max(1, (Emu ([Math]::Abs($X2 - $X1))))
  $cy = [Math]::Max(1, (Emu ([Math]::Abs($Y2 - $Y1))))
  $flipH = ""
  $flipV = ""
  if ($X2 -lt $X1) { $flipH = " flipH=""1""" }
  if ($Y2 -lt $Y1) { $flipV = " flipV=""1""" }
  $w = [Math]::Max(1, $Width) * 12700
  $head = ""
  if ($Arrow) { $head = "<a:headEnd type=""triangle""/>" }
  $xml = @"
<p:cxnSp>
  <p:nvCxnSpPr><p:cNvPr id="$id" name="Connector $id"/><p:cNvCxnSpPr/><p:nvPr/></p:nvCxnSpPr>
  <p:spPr><a:xfrm$flipH$flipV><a:off x="$x" y="$y"/><a:ext cx="$cx" cy="$cy"/></a:xfrm><a:prstGeom prst="line"><a:avLst/></a:prstGeom><a:ln w="$w"><a:solidFill><a:srgbClr val="$Color"/></a:solidFill>$head</a:ln></p:spPr>
</p:cxnSp>
"@
  $script:Shapes.Add($xml) | Out-Null
}

function Add-Header {
  param([string]$Title, [string]$Subtitle = "")
  Add-Shape -X 0 -Y 0 -W $SlideW -H 0.14 -Fill $Colors.Teal -Line "none" -Geom "rect"
  Add-Text -X 0.55 -Y 0.32 -W 10.6 -H 0.48 -Lines @($Title) -FontSize 24 -Color $Colors.Ink -Bold $true
  if ($Subtitle.Length -gt 0) {
    Add-Text -X 0.58 -Y 0.82 -W 10.2 -H 0.26 -Lines @($Subtitle) -FontSize 9 -Color $Colors.Muted
  }
  Add-Text -X 11.55 -Y 0.36 -W 1.2 -H 0.28 -Lines @("MCA Project") -FontSize 9 -Color $Colors.Muted -Align "r"
}

function Add-Footer {
  param([int]$Number)
  Add-Text -X 0.55 -Y 7.08 -W 5.0 -H 0.22 -Lines @("Scalable Ecommerce Platform | Final Project") -FontSize 8 -Color $Colors.Muted
  Add-Text -X 12.12 -Y 7.08 -W 0.6 -H 0.22 -Lines @([string]$Number) -FontSize 8 -Color $Colors.Muted -Align "r"
}

function Add-Stat {
  param([double]$X, [double]$Y, [double]$W, [string]$Value, [string]$Label, [string]$Accent)
  Add-Shape -X $X -Y $Y -W $W -H 1.05 -Fill "FFFFFF" -Line "E2E8F0" -LineWidth 1 -Geom "roundRect"
  Add-Shape -X ($X + 0.12) -Y ($Y + 0.16) -W 0.1 -H 0.72 -Fill $Accent -Line "none" -Geom "rect"
  Add-Text -X ($X + 0.32) -Y ($Y + 0.18) -W ($W - 0.5) -H 0.32 -Lines @($Value) -FontSize 23 -Color $Accent -Bold $true
  Add-Text -X ($X + 0.34) -Y ($Y + 0.58) -W ($W - 0.5) -H 0.28 -Lines @($Label) -FontSize 10 -Color $Colors.Muted
}

function Add-Card {
  param([double]$X, [double]$Y, [double]$W, [double]$H, [string]$Title, [string[]]$Lines, [string]$Accent = "2563EB", [string]$Fill = "FFFFFF")
  Add-Shape -X $X -Y $Y -W $W -H $H -Fill $Fill -Line "E2E8F0" -LineWidth 1 -Geom "roundRect"
  Add-Shape -X ($X + 0.16) -Y ($Y + 0.18) -W 0.12 -H ($H - 0.36) -Fill $Accent -Line "none" -Geom "rect"
  Add-Text -X ($X + 0.42) -Y ($Y + 0.18) -W ($W - 0.62) -H 0.28 -Lines @($Title) -FontSize 13 -Color $Colors.Ink -Bold $true
  Add-Text -X ($X + 0.42) -Y ($Y + 0.55) -W ($W - 0.62) -H ($H - 0.7) -Lines $Lines -FontSize 9 -Color $Colors.Muted
}

function Add-Placeholder {
  param([double]$X, [double]$Y, [double]$W, [double]$H, [string]$Text)
  Add-Shape -X $X -Y $Y -W $W -H $H -Fill "F8FAFC" -Line "94A3B8" -LineWidth 1 -Geom "roundRect" -Dash $true
  Add-Text -X ($X + 0.25) -Y ($Y + ($H / 2) - 0.22) -W ($W - 0.5) -H 0.44 -Lines @($Text) -FontSize 13 -Color $Colors.Muted -Align "c" -Anchor "mid"
}

function Add-Step {
  param([int]$N, [double]$X, [double]$Y, [double]$W, [string]$Title, [string]$Detail, [string]$Accent = "2563EB")
  Add-Shape -X $X -Y $Y -W 0.42 -H 0.42 -Text @([string]$N) -Fill $Accent -Line "none" -Geom "ellipse" -FontSize 12 -FontColor "FFFFFF" -Bold $true -Align "c"
  Add-Text -X ($X + 0.55) -Y ($Y - 0.02) -W $W -H 0.24 -Lines @($Title) -FontSize 12 -Color $Colors.Ink -Bold $true
  Add-Text -X ($X + 0.55) -Y ($Y + 0.25) -W $W -H 0.33 -Lines @($Detail) -FontSize 8 -Color $Colors.Muted
}

function Add-FlowBox {
  param([double]$X, [double]$Y, [double]$W, [string]$Text, [string]$Fill, [string]$Accent)
  Add-Shape -X $X -Y $Y -W $W -H 0.56 -Text @($Text) -Fill $Fill -Line $Accent -LineWidth 1 -Geom "roundRect" -FontSize 10 -FontColor $Colors.Ink -Bold $true -Align "c"
}

function New-Slide {
  param([string]$Title, [string]$Subtitle, [scriptblock]$Body)
  $script:ShapeId = 1
  $script:Shapes = New-Object System.Collections.Generic.List[string]
  Add-Header -Title $Title -Subtitle $Subtitle
  & $Body
  Add-Footer -Number ($script:Slides.Count + 1)
  $shapeXml = $script:Shapes -join "`n"
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg><p:bgPr><a:solidFill><a:srgbClr val="F8FAFC"/></a:solidFill><a:effectLst/></p:bgPr></p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>
      $shapeXml
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>
"@
}

$script:Slides = New-Object System.Collections.Generic.List[string]

$script:Slides.Add((New-Slide -Title "Scalable Backend Development for High-Performance E-Commerce Website Platform" -Subtitle "MCA Final Year Project Presentation" -Body {
  Add-Shape -X 0 -Y 0.14 -W $SlideW -H 2.1 -Fill $Colors.Navy -Line "none" -Geom "rect"
  Add-Shape -X 0 -Y 2.05 -W $SlideW -H 0.26 -Fill $Colors.Teal -Line "none" -Geom "rect"
  Add-Text -X 0.72 -Y 0.68 -W 11.2 -H 0.76 -Lines @("Scalable Backend Development for High-Performance E-Commerce Website Platform") -FontSize 31 -Color "FFFFFF" -Bold $true
  Add-Text -X 0.76 -Y 1.52 -W 9.5 -H 0.32 -Lines @("Full-stack microservices ecommerce platform with buyer, seller, analytics, and superadmin experiences") -FontSize 13 -Color "DDEBFF"
  Add-Stat -X 0.82 -Y 2.85 -W 2.75 -Value "14" -Label "Go backend services" -Accent $Colors.Teal
  Add-Stat -X 3.9 -Y 2.85 -W 2.75 -Value "4" -Label "React frontend apps" -Accent $Colors.Blue
  Add-Stat -X 6.98 -Y 2.85 -W 2.75 -Value "99" -Label "REST API endpoints" -Accent $Colors.Amber
  Add-Stat -X 10.06 -Y 2.85 -W 2.45 -Value "5+" -Label "data/infra systems" -Accent $Colors.Violet
  Add-Card -X 0.9 -Y 4.52 -W 5.5 -H 1.2 -Title "Submitted By" -Lines @("Parag Gulati", "Master of Computer Applications", "Final Year Project | 2026") -Accent $Colors.Teal
  Add-Card -X 6.75 -Y 4.52 -W 5.0 -H 1.2 -Title "Project Source" -Lines @("Repository code, final report, API catalog, docs, database schemas, runbooks, and frontend routes") -Accent $Colors.Blue
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Project Introduction" -Subtitle "What the platform is designed to demonstrate" -Body {
  Add-Text -X 0.75 -Y 1.35 -W 7.6 -H 0.65 -Lines @("A modular ecommerce platform built around independent services, role-specific frontends, secure API access, and service-owned data.") -FontSize 20 -Color $Colors.Ink -Bold $true
  Add-Card -X 0.8 -Y 2.35 -W 3.7 -H 1.12 -Title "Buyer Experience" -Lines @("Browse catalog", "Search and categories", "Wishlist, cart, checkout", "Orders and profile") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.85 -Y 2.35 -W 3.7 -H 1.12 -Title "Seller Experience" -Lines @("Product manager", "Order fulfillment", "Coupons and campaigns", "Analytics, team, audit") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 8.9 -Y 2.35 -W 3.7 -H 1.12 -Title "Admin Experience" -Lines @("Session analytics", "User/seller controls", "Payment/refund review", "Settings and audit logs") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 0.8 -Y 4.08 -W 5.7 -H 1.45 -Title "Architecture Direction" -Lines @("Public REST APIs through API Gateway", "gRPC/protobuf contracts for typed internal communication", "Events for asynchronous indexing, notification, and analytics workflows") -Accent $Colors.Navy
  Add-Card -X 6.85 -Y 4.08 -W 5.7 -H 1.45 -Title "Academic Value" -Lines @("Covers requirement analysis, architecture, database design, backend, frontend, DevOps, testing, limitations, and future scope") -Accent $Colors.Amber -Fill $Colors.SoftAmber
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Problem Statement" -Subtitle "Why the project is required" -Body {
  Add-Text -X 0.8 -Y 1.28 -W 11.2 -H 0.48 -Lines @("A growing ecommerce system must support shopping, seller operations, payments, analytics, and administration without becoming a tightly coupled monolith.") -FontSize 18 -Color $Colors.Ink -Bold $true
  Add-Card -X 0.8 -Y 2.15 -W 2.85 -H 1.38 -Title "Scaling Pressure" -Lines @("Catalog, search, checkout, and analytics have different load patterns.") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 3.95 -Y 2.15 -W 2.85 -H 1.38 -Title "Tight Coupling" -Lines @("A shared database and one codebase make changes risky.") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Card -X 7.1 -Y 2.15 -W 2.85 -H 1.38 -Title "Operational Gaps" -Lines @("Admins need visibility into users, sessions, payments, and audit history.") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 10.25 -Y 2.15 -W 2.25 -H 1.38 -Title "Reliability" -Lines @("Checkout and payment flows require idempotency and webhook handling.") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Shape -X 1.05 -Y 4.33 -W 11.3 -H 1.28 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.38 -Y 4.62 -W 10.5 -H 0.55 -Lines @("The project solves this by separating business areas into services, routing requests through an API Gateway, assigning data ownership per service, and documenting the full platform for academic evaluation.") -FontSize 16 -Color $Colors.Ink -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Objectives" -Subtitle "Main implementation goals" -Body {
  Add-Card -X 0.75 -Y 1.35 -W 3.75 -H 1.1 -Title "Modular Backend" -Lines @("Create independent services for auth, users, catalog, cart, wishlist, order, payment, search, CMS, analytics, notification, recommendation, and superadmin.") -Accent $Colors.Teal
  Add-Card -X 4.8 -Y 1.35 -W 3.75 -H 1.1 -Title "Secure Access" -Lines @("Use JWT, refresh tokens, OTP, role mapping, protected routes, and RBAC for buyer, seller, admin, and superadmin flows.") -Accent $Colors.Blue
  Add-Card -X 8.85 -Y 1.35 -W 3.75 -H 1.1 -Title "Role-Based UIs" -Lines @("Provide separate React frontends for buyers, sellers, session analytics admins, and superadmins.") -Accent $Colors.Violet
  Add-Card -X 0.75 -Y 3.0 -W 3.75 -H 1.1 -Title "Service-Owned Data" -Lines @("Use MySQL, MongoDB, Redis, Typesense, and queues according to each workload.") -Accent $Colors.Amber
  Add-Card -X 4.8 -Y 3.0 -W 3.75 -H 1.1 -Title "Reliable Checkout" -Lines @("Coordinate cart, product inventory, orders, payment intent, webhook status, and notifications.") -Accent $Colors.Green
  Add-Card -X 8.85 -Y 3.0 -W 3.75 -H 1.1 -Title "Deploy and Validate" -Lines @("Support Docker Compose, Kubernetes/Kustomize manifests, CI checks, health endpoints, logs, metrics, and tracing.") -Accent $Colors.Navy
  Add-Shape -X 0.85 -Y 5.05 -W 11.5 -H 0.72 -Fill $Colors.Navy -Line "none" -Geom "roundRect"
  Add-Text -X 1.12 -Y 5.26 -W 10.9 -H 0.28 -Lines @("Objective summary: build a college-submission-ready proof of modern ecommerce system design, implementation, and documentation.") -FontSize 13 -Color "FFFFFF" -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Scope of Project" -Subtitle "Features included and boundaries documented" -Body {
  Add-Card -X 0.75 -Y 1.28 -W 3.75 -H 1.55 -Title "Buyer Scope" -Lines @("Login/register/OTP/reset pages", "Home, categories, deals, search", "Product detail, wishlist, cart", "Checkout, payment result, orders", "Profile, addresses, notifications") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.8 -Y 1.28 -W 3.75 -H 1.55 -Title "Seller Scope" -Lines @("Seller login/session", "Product list and editor", "Orders and fulfillment", "Coupons, campaigns, offers", "Revenue analytics, team, audit") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 8.85 -Y 1.28 -W 3.75 -H 1.55 -Title "Admin Scope" -Lines @("Session analytics dashboard", "Superadmin users and sellers", "Orders, payments, refunds", "Settings and audit logs", "Search route is a placeholder") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 0.75 -Y 3.42 -W 5.75 -H 1.35 -Title "Engineering Scope" -Lines @("14 backend services, 99 API catalog endpoints, database schema files, protobuf contracts, Docker Compose, Kubernetes manifests, observability config, runbooks, and CI.") -Accent $Colors.Navy
  Add-Card -X 6.85 -Y 3.42 -W 5.75 -H 1.35 -Title "Clearly Not Claimed" -Lines @("No fake credentials, no fake screenshots, no production payment keys, no proven live production deployment, and no full browser E2E suite found.") -Accent $Colors.Red -Fill $Colors.SoftRed
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Technology Stack" -Subtitle "Tools identified from project source" -Body {
  Add-Card -X 0.8 -Y 1.25 -W 3.6 -H 4.25 -Title "Frontend" -Lines @("React", "TypeScript", "Vite", "React Router", "TanStack Query", "Zustand", "React Hook Form", "Zod", "Tailwind CSS", "Lucide icons") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.85 -Y 1.25 -W 3.6 -H 4.25 -Title "Backend and APIs" -Lines @("Go microservices", "REST through API Gateway", "gRPC/protobuf contracts", "gRPC-Web facade support", "JWT and RBAC", "Request validation", "Health endpoints", "Service runbooks") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 8.9 -Y 1.25 -W 3.6 -H 4.25 -Title "Data, Infra, DevOps" -Lines @("MySQL", "MongoDB", "Redis", "Typesense", "RabbitMQ/Kafka", "Docker Compose", "Kubernetes/Kustomize", "Prometheus", "OpenTelemetry", "Jaeger") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Shape -X 1.05 -Y 5.95 -W 11.3 -H 0.38 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.25 -Y 6.03 -W 10.9 -H 0.16 -Lines @("Design principle: each major business capability owns its data and exposes behavior through APIs, not direct cross-service database access.") -FontSize 9 -Color $Colors.Muted -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "System Architecture" -Subtitle "High-level request and service layout" -Body {
  Add-Shape -X 0.75 -Y 1.3 -W 2.7 -H 3.75 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.0 -Y 1.55 -W 2.2 -H 0.25 -Lines @("Frontend Clients") -FontSize 13 -Color $Colors.Ink -Bold $true -Align "c"
  Add-FlowBox -X 1.05 -Y 2.0 -W 2.1 -Text "User App" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-FlowBox -X 1.05 -Y 2.72 -W 2.1 -Text "Seller Dashboard" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-FlowBox -X 1.05 -Y 3.44 -W 2.1 -Text "Analytics Dashboard" -Fill $Colors.SoftViolet -Accent $Colors.Violet
  Add-FlowBox -X 1.05 -Y 4.16 -W 2.1 -Text "Superadmin Panel" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-FlowBox -X 4.55 -Y 2.72 -W 2.05 -Text "API Gateway" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Line -X1 3.45 -Y1 3.0 -X2 4.55 -Y2 3.0 -Color $Colors.Navy -Width 2
  Add-Shape -X 7.45 -Y 1.3 -W 4.95 -H 3.75 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 7.75 -Y 1.55 -W 4.35 -H 0.25 -Lines @("Business Services") -FontSize 13 -Color $Colors.Ink -Bold $true -Align "c"
  $services1 = @("Auth", "User", "Product", "Cart", "Wishlist", "Search", "Order", "Payment", "CMS", "Session", "Notification", "Recommendation", "Superadmin")
  $sx = 7.75; $sy = 1.95; $i = 0
  foreach ($svc in $services1) {
    $col = $i % 3; $row = [Math]::Floor($i / 3)
    Add-Shape -X ($sx + ($col * 1.45)) -Y ($sy + ($row * 0.56)) -W 1.25 -H 0.38 -Text @($svc) -Fill $Colors.Light -Line "E2E8F0" -Geom "roundRect" -FontSize 8 -FontColor $Colors.Ink -Bold $true -Align "c"
    $i++
  }
  Add-Line -X1 6.6 -Y1 3.0 -X2 7.45 -Y2 3.0 -Color $Colors.Navy -Width 2
  Add-Shape -X 1.25 -Y 5.55 -W 10.85 -H 0.75 -Fill $Colors.Navy -Line "none" -Geom "roundRect"
  Add-Text -X 1.55 -Y 5.73 -W 10.25 -H 0.27 -Lines @("Data Stores and Async Infrastructure: MySQL | MongoDB | Redis | Typesense | RabbitMQ/Kafka | Mailpit | Prometheus | Jaeger") -FontSize 12 -Color "FFFFFF" -Bold $true -Align "c"
  Add-Line -X1 9.9 -Y1 5.05 -X2 7.0 -Y2 5.55 -Color $Colors.Muted -Width 1
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Modules and Services Overview" -Subtitle "Backend service inventory from the project folder" -Body {
  Add-Card -X 0.75 -Y 1.2 -W 3.0 -H 1.65 -Title "Access and Identity" -Lines @("API Gateway", "Auth Service", "User Service", "JWT, refresh tokens, OTP", "Role mapping and protected routes") -Accent $Colors.Navy
  Add-Card -X 3.95 -Y 1.2 -W 3.0 -H 1.65 -Title "Shopping Core" -Lines @("Product Service", "Cart Service", "Wishlist Service", "Search Service", "Recommendation Service") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 7.15 -Y 1.2 -W 2.75 -H 1.65 -Title "Order to Payment" -Lines @("Order Service", "Payment Service", "Inventory coordination", "Payment intents, refunds, webhooks") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 10.1 -Y 1.2 -W 2.45 -H 1.65 -Title "Operations" -Lines @("CMS Service", "Session Service", "Notification Service", "Superadmin Service") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Shape -X 0.85 -Y 3.55 -W 11.55 -H 1.55 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.15 -Y 3.85 -W 2.4 -H 0.28 -Lines @("Route Ownership") -FontSize 14 -Color $Colors.Ink -Bold $true
  Add-Stat -X 3.7 -Y 3.78 -W 1.75 -Value "42" -Label "admin endpoints" -Accent $Colors.Violet
  Add-Stat -X 5.7 -Y 3.78 -W 1.75 -Value "24" -Label "buyer endpoints" -Accent $Colors.Teal
  Add-Stat -X 7.7 -Y 3.78 -W 1.75 -Value "16" -Label "seller endpoints" -Accent $Colors.Blue
  Add-Stat -X 9.7 -Y 3.78 -W 1.75 -Value "13" -Label "public endpoints" -Accent $Colors.Green
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Database Overview" -Subtitle "Storage selected according to workload" -Body {
  Add-Card -X 0.75 -Y 1.25 -W 3.7 -H 3.75 -Title "MySQL: Transactional Data" -Lines @("Auth accounts and credentials", "User profiles and addresses", "Orders and order items", "Payments and refunds", "CMS coupons/campaigns", "Superadmin RBAC and audit logs") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 4.8 -Y 1.25 -W 3.7 -H 3.75 -Title "MongoDB: Documents and Events" -Lines @("Products and variants", "Carts and wishlists", "Session events and journeys", "Recommendations", "Notification templates and deliveries") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 8.85 -Y 1.25 -W 3.7 -H 1.72 -Title "Redis: Fast Short-Lived State" -Lines @("Rate limits", "Sessions and OTP counters", "Cart cache", "Active sessions", "Distributed locks where needed") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Card -X 8.85 -Y 3.25 -W 3.7 -H 1.75 -Title "Typesense and Queues" -Lines @("Typesense for product search, facets, autocomplete", "RabbitMQ/Kafka for events, indexing, notifications, analytics") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Text -X 1.0 -Y 5.58 -W 11.3 -H 0.42 -Lines @("Database rule: services do not directly read or write another service database. They call APIs or consume events.") -FontSize 14 -Color $Colors.Ink -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "API and Backend Flow" -Subtitle "Gateway-first request flow and checkout coordination" -Body {
  $x0 = 0.75; $y = 1.55
  Add-FlowBox -X $x0 -Y $y -W 1.6 -Text "React App" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 2.35 -Y1 ($y + 0.28) -X2 3.1 -Y2 ($y + 0.28) -Color $Colors.Navy
  Add-FlowBox -X 3.1 -Y $y -W 1.75 -Text "API Gateway" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Line -X1 4.85 -Y1 ($y + 0.28) -X2 5.6 -Y2 ($y + 0.28) -Color $Colors.Navy
  Add-FlowBox -X 5.6 -Y $y -W 1.55 -Text "Auth" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 7.15 -Y1 ($y + 0.28) -X2 7.9 -Y2 ($y + 0.28) -Color $Colors.Navy
  Add-FlowBox -X 7.9 -Y $y -W 1.55 -Text "Order" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Line -X1 9.45 -Y1 ($y + 0.28) -X2 10.2 -Y2 ($y + 0.28) -Color $Colors.Navy
  Add-FlowBox -X 10.2 -Y $y -W 1.85 -Text "Payment" -Fill $Colors.SoftRed -Accent $Colors.Red
  Add-Card -X 0.75 -Y 2.72 -W 3.55 -H 2.25 -Title "Gateway Responsibilities" -Lines @("Route catalog", "JWT verification", "Role checks", "Request validation", "Rate limiting", "CORS, logs, metrics, traces", "Error mapping") -Accent $Colors.Navy
  Add-Card -X 4.68 -Y 2.72 -W 3.55 -H 2.25 -Title "Checkout Flow" -Lines @("Get cart", "Validate user and address", "Reserve inventory", "Create pending order", "Create payment intent", "Publish order/payment events") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 8.6 -Y 2.72 -W 3.55 -H 2.25 -Title "Payment Truth Source" -Lines @("Frontend callback is not final", "Provider webhook updates payment", "Order status changes after backend confirmation", "Refunds are admin-controlled") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Text -X 1.1 -Y 5.72 -W 11.0 -H 0.25 -Lines @("API catalog evidence: 99 REST endpoints mapped across public, buyer, seller, admin, superadmin, and webhook auth levels.") -FontSize 11 -Color $Colors.Muted -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Frontend Overview" -Subtitle "Implemented route surfaces" -Body {
  Add-Card -X 0.75 -Y 1.18 -W 2.95 -H 2.05 -Title "User App :3000" -Lines @("Home, categories, deals", "Search and product detail", "Login, signup, OTP, reset", "Wishlist, cart, checkout", "Profile, addresses, orders, notifications") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 3.95 -Y 1.18 -W 2.95 -H 2.05 -Title "Seller Dashboard :3001" -Lines @("Products and editor", "Orders and fulfillment", "Offers, coupons, campaigns", "Revenue analytics", "Team and audit activity") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 7.15 -Y 1.18 -W 2.95 -H 2.05 -Title "Analytics :3002" -Lines @("Overview and live sessions", "Journey explorer", "Funnels and heatmaps", "Cohorts and reports", "Privacy controls") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 10.35 -Y 1.18 -W 2.25 -H 2.05 -Title "Superadmin :3003" -Lines @("Users and sellers", "Orders", "Payments/refunds", "Sessions", "Settings, audit logs", "Search placeholder") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Placeholder -X 0.9 -Y 3.75 -W 3.45 -H 1.55 -Text "Screenshot to be added: User App Home / Product Listing"
  Add-Placeholder -X 4.95 -Y 3.75 -W 3.45 -H 1.55 -Text "Screenshot to be added: Seller Dashboard"
  Add-Placeholder -X 9.0 -Y 3.75 -W 3.0 -H 1.55 -Text "Screenshot to be added: Admin / Analytics UI"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Step-by-Step User Guide: Access Frontend" -Subtitle "Local run path from repository runbooks" -Body {
  Add-Step -N 1 -X 0.9 -Y 1.35 -W 5.0 -Title "Start the local stack" -Detail "Run docker compose up -d --build from the repository root." -Accent $Colors.Navy
  Add-Step -N 2 -X 0.9 -Y 2.12 -W 5.0 -Title "Verify gateway health" -Detail "Open /health/live and /health/ready on http://localhost:8080." -Accent $Colors.Blue
  Add-Step -N 3 -X 0.9 -Y 2.89 -W 5.0 -Title "Open the buyer frontend" -Detail "Use http://localhost:3000 for the user-facing app." -Accent $Colors.Teal
  Add-Step -N 4 -X 0.9 -Y 3.66 -W 5.0 -Title "Register or log in" -Detail "Signup, login, OTP, forgot password, and reset password pages exist in the user app." -Accent $Colors.Violet
  Add-Step -N 5 -X 0.9 -Y 4.43 -W 5.0 -Title "Use public and protected pages" -Detail "Catalog/search pages are public; wishlist, cart, checkout, and account pages are protected." -Accent $Colors.Amber
  Add-Card -X 7.05 -Y 1.42 -W 5.15 -H 1.6 -Title "Important Access Note" -Lines @("Buyer seed credentials were not found in the local access guide.", "Seller, analytics admin, and superadmin demo credentials exist and are documented there.", "No fake buyer credentials are included in this presentation.") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Placeholder -X 7.05 -Y 3.45 -W 5.15 -H 1.85 -Text "Screenshot to be added: User App Login or Signup Page"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Client/User Journey: Authentication and Discovery" -Subtitle "Normal buyer flow, based on implemented routes" -Body {
  Add-FlowBox -X 0.8 -Y 1.45 -W 1.55 -Text "Signup/Login" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 2.35 -Y1 1.73 -X2 3.0 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 3.0 -Y 1.45 -W 1.25 -Text "OTP" -Fill $Colors.SoftViolet -Accent $Colors.Violet
  Add-Line -X1 4.25 -Y1 1.73 -X2 4.9 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 4.9 -Y 1.45 -W 1.55 -Text "Home" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 6.45 -Y1 1.73 -X2 7.1 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 7.1 -Y 1.45 -W 1.65 -Text "Search/Filter" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Line -X1 8.75 -Y1 1.73 -X2 9.4 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 9.4 -Y 1.45 -W 1.75 -Text "Product Detail" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Line -X1 11.15 -Y1 1.73 -X2 11.75 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 11.75 -Y 1.45 -W 0.95 -Text "Cart" -Fill $Colors.SoftRed -Accent $Colors.Red
  Add-Card -X 0.8 -Y 2.72 -W 5.55 -H 1.4 -Title "Discovery Features Present" -Lines @("Home page, category listing, category route, deals page, search results, product detail page, sort and filter URL state, wishlist and cart actions.") -Accent $Colors.Teal
  Add-Card -X 6.75 -Y 2.72 -W 5.55 -H 1.4 -Title "Authentication Features Present" -Lines @("Login, signup, OTP page, forgot password, reset password, auth store, protected route wrapper, token/session handling utilities.") -Accent $Colors.Blue
  Add-Placeholder -X 1.15 -Y 4.55 -W 5.0 -H 1.2 -Text "Screenshot to be added: Search / Category Filter"
  Add-Placeholder -X 7.15 -Y 4.55 -W 5.0 -H 1.2 -Text "Screenshot to be added: Product Detail"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Client/User Journey: Cart, Checkout and Payment" -Subtitle "Protected buyer workflow" -Body {
  Add-FlowBox -X 0.85 -Y 1.38 -W 1.35 -Text "Cart" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 2.2 -Y1 1.66 -X2 2.85 -Y2 1.66 -Color $Colors.Navy
  Add-FlowBox -X 2.85 -Y 1.38 -W 1.75 -Text "Qty/Coupon" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 4.6 -Y1 1.66 -X2 5.25 -Y2 1.66 -Color $Colors.Navy
  Add-FlowBox -X 5.25 -Y 1.38 -W 1.45 -Text "Address" -Fill $Colors.SoftViolet -Accent $Colors.Violet
  Add-Line -X1 6.7 -Y1 1.66 -X2 7.35 -Y2 1.66 -Color $Colors.Navy
  Add-FlowBox -X 7.35 -Y 1.38 -W 1.55 -Text "Review" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Line -X1 8.9 -Y1 1.66 -X2 9.55 -Y2 1.66 -Color $Colors.Navy
  Add-FlowBox -X 9.55 -Y 1.38 -W 1.55 -Text "Payment" -Fill $Colors.SoftRed -Accent $Colors.Red
  Add-Line -X1 11.1 -Y1 1.66 -X2 11.75 -Y2 1.66 -Color $Colors.Navy
  Add-FlowBox -X 11.75 -Y 1.38 -W 0.95 -Text "Result" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Card -X 0.85 -Y 2.58 -W 3.55 -H 1.6 -Title "Cart Service Flow" -Lines @("Get cart", "Add item", "Update item quantity", "Remove item", "Merge guest cart", "Preview coupon") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.85 -Y 2.58 -W 3.55 -H 1.6 -Title "Order Service Flow" -Lines @("Create order from cart", "Use idempotency key", "Store order items", "Track order status", "List buyer orders") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 8.85 -Y 2.58 -W 3.55 -H 1.6 -Title "Payment Flow" -Lines @("Create payment intent", "Retry payment endpoint", "Provider webhook endpoint", "Refund is admin-side", "Local providers disabled by default") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Placeholder -X 1.15 -Y 4.78 -W 5.0 -H 1.15 -Text "Screenshot to be added: Cart Page"
  Add-Placeholder -X 7.15 -Y 4.78 -W 5.0 -H 1.15 -Text "Screenshot to be added: Checkout Page"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Client/User Journey: Profile and Notifications" -Subtitle "Post-purchase and account management" -Body {
  Add-Card -X 0.8 -Y 1.25 -W 3.7 -H 1.5 -Title "Account Overview" -Lines @("Protected account layout", "Profile overview route", "Profile detail page", "Session-aware navigation") -Accent $Colors.Navy
  Add-Card -X 4.8 -Y 1.25 -W 3.7 -H 1.5 -Title "Address Book" -Lines @("List addresses", "Create address", "Update address", "Delete address", "Use address during checkout") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 8.8 -Y 1.25 -W 3.7 -H 1.5 -Title "Orders" -Lines @("Order list page", "Order detail page", "Cancel order API exists", "Payment result pages") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 1.35 -Y 3.35 -W 4.7 -H 1.5 -Title "Notification Preferences" -Lines @("User app route /account/notifications", "Notification service schemas", "Email/SMS/push/marketing preference model") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Placeholder -X 7.15 -Y 3.25 -W 4.7 -H 1.7 -Text "Screenshot to be added: Profile / Orders / Notification Preferences"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Seller Dashboard Journey" -Subtitle "Seller/admin flow exists in code and runbook" -Body {
  Add-Card -X 0.75 -Y 1.25 -W 3.2 -H 1.45 -Title "Access" -Lines @("URL: http://localhost:3001", "Local demo login is documented in LOCAL_ACCESS_GUIDE.md", "Role/context: seller") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-FlowBox -X 4.4 -Y 1.45 -W 1.4 -Text "Login" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 5.8 -Y1 1.73 -X2 6.35 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 6.35 -Y 1.45 -W 1.55 -Text "Products" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 7.9 -Y1 1.73 -X2 8.45 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 8.45 -Y 1.45 -W 1.45 -Text "Orders" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Line -X1 9.9 -Y1 1.73 -X2 10.45 -Y2 1.73 -Color $Colors.Navy
  Add-FlowBox -X 10.45 -Y 1.45 -W 1.65 -Text "Analytics" -Fill $Colors.SoftViolet -Accent $Colors.Violet
  Add-Card -X 0.75 -Y 3.05 -W 3.7 -H 1.7 -Title "Product Management" -Lines @("Product list", "New product editor", "Edit product route", "Category/product APIs through gateway") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.8 -Y 3.05 -W 3.7 -H 1.7 -Title "Commercial Tools" -Lines @("Order list and detail", "Fulfillment update", "Offers page", "Coupon editor", "Campaign create page") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 8.85 -Y 3.05 -W 3.7 -H 1.7 -Title "Operations" -Lines @("Revenue analytics", "Team members", "Activity/audit page", "CMS service audit logs") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Placeholder -X 1.35 -Y 5.35 -W 10.3 -H 0.8 -Text "Screenshot to be added: Seller Product Manager / Orders / Analytics"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Session Analytics Dashboard Journey" -Subtitle "Operations analytics flow" -Body {
  Add-Card -X 0.75 -Y 1.25 -W 3.25 -H 1.45 -Title "Access" -Lines @("URL: http://localhost:3002", "Local demo operations admin is documented", "Requires admin-level role") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 4.35 -Y 1.25 -W 3.25 -H 1.45 -Title "Live Monitoring" -Lines @("Overview", "Live sessions", "Active users", "Events per minute", "Device/channel filters") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 7.95 -Y 1.25 -W 4.3 -H 1.45 -Title "Analysis Modules" -Lines @("Journey explorer, funnel analysis, heatmaps, cohort retention, report export, scheduled reports, privacy controls") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-FlowBox -X 1.0 -Y 3.25 -W 1.25 -Text "Live" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 2.25 -Y1 3.53 -X2 3.0 -Y2 3.53 -Color $Colors.Navy
  Add-FlowBox -X 3.0 -Y 3.25 -W 1.4 -Text "Journey" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 4.4 -Y1 3.53 -X2 5.15 -Y2 3.53 -Color $Colors.Navy
  Add-FlowBox -X 5.15 -Y 3.25 -W 1.35 -Text "Funnels" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Line -X1 6.5 -Y1 3.53 -X2 7.25 -Y2 3.53 -Color $Colors.Navy
  Add-FlowBox -X 7.25 -Y 3.25 -W 1.45 -Text "Heatmaps" -Fill $Colors.SoftRed -Accent $Colors.Red
  Add-Line -X1 8.7 -Y1 3.53 -X2 9.45 -Y2 3.53 -Color $Colors.Navy
  Add-FlowBox -X 9.45 -Y 3.25 -W 1.25 -Text "Reports" -Fill $Colors.SoftViolet -Accent $Colors.Violet
  Add-Line -X1 10.7 -Y1 3.53 -X2 11.45 -Y2 3.53 -Color $Colors.Navy
  Add-FlowBox -X 11.45 -Y 3.25 -W 1.15 -Text "Privacy" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Placeholder -X 1.25 -Y 4.75 -W 10.75 -H 1.05 -Text "Screenshot to be added: Analytics Overview / Funnel / Heatmap"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Superadmin Journey" -Subtitle "Platform control plane flow" -Body {
  Add-Card -X 0.75 -Y 1.2 -W 3.0 -H 1.45 -Title "Access" -Lines @("URL: http://localhost:3003", "Local demo superadmin is documented", "Role: superadmin") -Accent $Colors.Navy
  Add-Card -X 4.0 -Y 1.2 -W 2.75 -H 1.45 -Title "User and Seller Controls" -Lines @("User list/detail", "User status updates", "Seller list/review", "Seller status updates") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 7.0 -Y 1.2 -W 2.75 -H 1.45 -Title "Orders and Payments" -Lines @("Order operations", "Order detail", "Payment operations", "Refund review") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 10.0 -Y 1.2 -W 2.55 -H 1.45 -Title "Governance" -Lines @("Session oversight", "Platform settings", "Admin audit logs", "Search route placeholder") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Shape -X 1.0 -Y 3.35 -W 11.2 -H 1.2 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.3 -Y 3.62 -W 10.6 -H 0.5 -Lines @("Every high-risk admin mutation is designed to record audit evidence. Superadmin APIs coordinate with user, order, payment, session, and platform setting services through controlled backend handlers.") -FontSize 14 -Color $Colors.Ink -Bold $true -Align "c"
  Add-Placeholder -X 1.25 -Y 5.05 -W 10.75 -H 0.95 -Text "Screenshot to be added: Superadmin Users / Payments / Audit Logs"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Deployment and Local Run Overview" -Subtitle "How the project is executed and operated locally" -Body {
  Add-Card -X 0.75 -Y 1.25 -W 3.7 -H 1.65 -Title "Local Source of Truth" -Lines @("Root docker-compose.yml", "Starts frontend apps, backend services, databases, queues, observability tools", "Runbook index documents startup order and checks") -Accent $Colors.Navy
  Add-Card -X 4.8 -Y 1.25 -W 3.7 -H 1.65 -Title "Main Local URLs" -Lines @("User App :3000", "Seller Dashboard :3001", "Analytics :3002", "Superadmin :3003", "Gateway :8080", "Mailpit, RabbitMQ, Prometheus, Jaeger") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 8.85 -Y 1.25 -W 3.7 -H 1.65 -Title "Deployment Assets" -Lines @("14 backend Dockerfiles", "Kubernetes/Kustomize manifests", "Namespaces: edge, core, data, jobs, observability", "Secrets examples and config maps") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-FlowBox -X 1.1 -Y 3.9 -W 2.0 -Text "docker compose" -Fill "FFFFFF" -Accent $Colors.Navy
  Add-Line -X1 3.1 -Y1 4.18 -X2 4.0 -Y2 4.18 -Color $Colors.Navy
  Add-FlowBox -X 4.0 -Y 3.9 -W 2.0 -Text "Gateway + Services" -Fill $Colors.SoftBlue -Accent $Colors.Blue
  Add-Line -X1 6.0 -Y1 4.18 -X2 6.9 -Y2 4.18 -Color $Colors.Navy
  Add-FlowBox -X 6.9 -Y 3.9 -W 2.0 -Text "Data Stores" -Fill $Colors.SoftTeal -Accent $Colors.Teal
  Add-Line -X1 8.9 -Y1 4.18 -X2 9.8 -Y2 4.18 -Color $Colors.Navy
  Add-FlowBox -X 9.8 -Y 3.9 -W 2.0 -Text "Health + Logs" -Fill $Colors.SoftAmber -Accent $Colors.Amber
  Add-Text -X 1.1 -Y 5.32 -W 11.0 -H 0.3 -Lines @("Command: docker compose up -d --build | Verify: http://localhost:8080/health/live and /health/ready") -FontSize 12 -Color $Colors.Ink -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Testing and Validation" -Subtitle "Evidence found in source and CI" -Body {
  Add-Stat -X 0.85 -Y 1.25 -W 2.65 -Value "305" -Label "backend Go test files" -Accent $Colors.Blue
  Add-Stat -X 3.85 -Y 1.25 -W 2.65 -Value "461" -Label "frontend test/spec files" -Accent $Colors.Teal
  Add-Stat -X 6.85 -Y 1.25 -W 2.65 -Value "136" -Label "service migration files" -Accent $Colors.Amber
  Add-Stat -X 9.85 -Y 1.25 -W 2.65 -Value "14" -Label "service Dockerfiles" -Accent $Colors.Violet
  Add-Card -X 0.85 -Y 3.05 -W 3.7 -H 1.65 -Title "Backend CI" -Lines @("go test ./...", "go vet", "gofmt check", "golangci-lint", "Docker build checks") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 4.85 -Y 3.05 -W 3.7 -H 1.65 -Title "Frontend CI" -Lines @("workspace tests", "lint", "typecheck", "build", "React Testing Library and Vitest evidence") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 8.85 -Y 3.05 -W 3.7 -H 1.65 -Title "Contract and Security" -Lines @("buf lint", "buf breaking check", "generated client verification", "Hadolint", "Trivy high/critical scan") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Shape -X 1.1 -Y 5.45 -W 11.05 -H 0.55 -Fill $Colors.SoftRed -Line "FCA5A5" -Geom "roundRect"
  Add-Text -X 1.3 -Y 5.59 -W 10.65 -H 0.22 -Lines @("Known test gap: dedicated full browser E2E automation was not found in the repository.") -FontSize 11 -Color $Colors.Red -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Challenges Faced" -Subtitle "Practical engineering problems addressed by the project" -Body {
  Add-Card -X 0.75 -Y 1.25 -W 3.75 -H 1.25 -Title "Service Boundaries" -Lines @("Separating responsibilities while avoiding direct cross-service database access.") -Accent $Colors.Navy
  Add-Card -X 4.8 -Y 1.25 -W 3.75 -H 1.25 -Title "Auth and Roles" -Lines @("Supporting buyer, seller, admin, and superadmin workflows with different route guards.") -Accent $Colors.Blue
  Add-Card -X 8.85 -Y 1.25 -W 3.75 -H 1.25 -Title "Checkout Consistency" -Lines @("Coordinating cart, inventory, order, payment intent, webhook, and refund behavior.") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 0.75 -Y 3.1 -W 3.75 -H 1.25 -Title "Search and Events" -Lines @("Keeping product search index and recommendations updated without blocking user requests.") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.8 -Y 3.1 -W 3.75 -H 1.25 -Title "Local Orchestration" -Lines @("Running many services, data systems, migrations, health checks, and frontends together.") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 8.85 -Y 3.1 -W 3.75 -H 1.25 -Title "Documentation Accuracy" -Lines @("Keeping report, API catalog, route code, runbooks, and limitations aligned.") -Accent $Colors.Red -Fill $Colors.SoftRed
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Limitations and Assumptions" -Subtitle "Honest boundaries for college viva" -Body {
  Add-Card -X 0.75 -Y 1.18 -W 3.75 -H 1.5 -Title "Local Demo Limits" -Lines @("Buyer seed credentials not found", "Payment providers disabled in local defaults", "Payment events endpoint blank by default") -Accent $Colors.Red -Fill $Colors.SoftRed
  Add-Card -X 4.8 -Y 1.18 -W 3.75 -H 1.5 -Title "Implementation Gaps" -Lines @("Superadmin search route is a placeholder", "Dedicated browser E2E tests not found", "Public password reset route mismatch noted in report") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 8.85 -Y 1.18 -W 3.75 -H 1.5 -Title "Deployment Limits" -Lines @("Production secrets are not committed", "Kubernetes manifests exist but complete production deployment is not proven", "Production incident runbooks need expansion") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Shape -X 0.95 -Y 3.42 -W 11.35 -H 1.12 -Fill "FFFFFF" -Line "E2E8F0" -Geom "roundRect"
  Add-Text -X 1.25 -Y 3.65 -W 10.75 -H 0.44 -Lines @("Screenshot assumption: Headless Chrome and Edge both exited with code 13 in this environment, so the deck uses explicit screenshot placeholders instead of fabricated images.") -FontSize 14 -Color $Colors.Ink -Bold $true -Align "c"
  Add-Text -X 1.25 -Y 5.25 -W 10.75 -H 0.3 -Lines @("Academic assumption: Student name is taken from the provided file name. Other college details should be filled by the student before final submission.") -FontSize 11 -Color $Colors.Muted -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Future Enhancements" -Subtitle "Next steps for stronger production readiness" -Body {
  Add-Card -X 0.75 -Y 1.2 -W 3.75 -H 1.45 -Title "Functional Enhancements" -Lines @("Add buyer seed/demo account", "Complete superadmin search module", "Configure sandbox payment provider", "Improve notification provider webhooks") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 4.8 -Y 1.2 -W 3.75 -H 1.45 -Title "Quality Enhancements" -Lines @("Add Playwright/Cypress E2E tests", "Add contract drift tests", "Capture final screenshots", "Add mobile visual validation") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 8.85 -Y 1.2 -W 3.75 -H 1.45 -Title "DevOps Enhancements" -Lines @("Production secret management", "Helm charts or complete overlays", "Backup/restore runbooks", "Incident response runbooks") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Card -X 1.25 -Y 3.45 -W 5.1 -H 1.45 -Title "Performance and Scale" -Lines @("Load testing for gateway, search, checkout, session ingestion, Redis, MongoDB, MySQL, and queues.") -Accent $Colors.Violet -Fill $Colors.SoftViolet
  Add-Card -X 7.0 -Y 3.45 -W 5.1 -H 1.45 -Title "Intelligence" -Lines @("Advance recommendation engine from rule-based ranking toward ML/vector personalization.") -Accent $Colors.Green -Fill $Colors.SoftTeal
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Conclusion" -Subtitle "Final academic summary" -Body {
  Add-Text -X 0.95 -Y 1.4 -W 11.4 -H 0.64 -Lines @("The project demonstrates a complete MCA-level full-stack ecommerce system design with microservices, multiple role-based frontends, service-owned databases, API Gateway routing, testing, deployment planning, and formal documentation.") -FontSize 20 -Color $Colors.Ink -Bold $true -Align "c"
  Add-Card -X 0.95 -Y 2.65 -W 3.55 -H 1.35 -Title "Technical Outcome" -Lines @("Modern backend and frontend architecture with clear separation of responsibilities.") -Accent $Colors.Blue -Fill $Colors.SoftBlue
  Add-Card -X 4.9 -Y 2.65 -W 3.55 -H 1.35 -Title "Learning Outcome" -Lines @("Practical exposure to Go, React, APIs, databases, DevOps, security, testing, and documentation.") -Accent $Colors.Teal -Fill $Colors.SoftTeal
  Add-Card -X 8.85 -Y 2.65 -W 3.55 -H 1.35 -Title "Viva Readiness" -Lines @("Includes objective, scope, architecture, modules, workflows, limitations, and future scope.") -Accent $Colors.Amber -Fill $Colors.SoftAmber
  Add-Shape -X 1.65 -Y 5.15 -W 10.0 -H 0.65 -Fill $Colors.Navy -Line "none" -Geom "roundRect"
  Add-Text -X 1.9 -Y 5.32 -W 9.5 -H 0.24 -Lines @("No unsupported modules, fake screenshots, fake credentials, or fake implementation details were added.") -FontSize 12 -Color "FFFFFF" -Bold $true -Align "c"
})) | Out-Null

$script:Slides.Add((New-Slide -Title "Thank You" -Subtitle "Questions and Answers" -Body {
  Add-Shape -X 0 -Y 0.14 -W $SlideW -H 2.15 -Fill $Colors.Navy -Line "none" -Geom "rect"
  Add-Text -X 1.0 -Y 0.82 -W 11.3 -H 0.7 -Lines @("Thank You") -FontSize 44 -Color "FFFFFF" -Bold $true -Align "c"
  Add-Text -X 1.0 -Y 1.58 -W 11.3 -H 0.36 -Lines @("Questions and Answers") -FontSize 19 -Color "DDEBFF" -Align "c"
  Add-Card -X 1.25 -Y 3.0 -W 4.9 -H 1.35 -Title "Project Title" -Lines @("Scalable Backend Development for High-Performance E-Commerce Website Platform") -Accent $Colors.Teal
  Add-Card -X 7.15 -Y 3.0 -W 4.9 -H 1.35 -Title "Presented By" -Lines @("Parag Gulati", "MCA Final Year Project") -Accent $Colors.Blue
  Add-Text -X 1.45 -Y 5.2 -W 10.45 -H 0.3 -Lines @("Prepared from actual repository code, docs, API catalog, database schemas, frontend routes, runbooks, and final report.") -FontSize 11 -Color $Colors.Muted -Align "c"
})) | Out-Null

function ThemeXml {
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="MCA Ecommerce Theme">
  <a:themeElements>
    <a:clrScheme name="MCA">
      <a:dk1><a:srgbClr val="132033"/></a:dk1>
      <a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="12355B"/></a:dk2>
      <a:lt2><a:srgbClr val="F8FAFC"/></a:lt2>
      <a:accent1><a:srgbClr val="0F766E"/></a:accent1>
      <a:accent2><a:srgbClr val="2563EB"/></a:accent2>
      <a:accent3><a:srgbClr val="F59E0B"/></a:accent3>
      <a:accent4><a:srgbClr val="6D28D9"/></a:accent4>
      <a:accent5><a:srgbClr val="16A34A"/></a:accent5>
      <a:accent6><a:srgbClr val="DC2626"/></a:accent6>
      <a:hlink><a:srgbClr val="2563EB"/></a:hlink>
      <a:folHlink><a:srgbClr val="6D28D9"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="Aptos">
      <a:majorFont><a:latin typeface="Aptos Display"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="Aptos"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="MCA">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="6350" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
        <a:ln w="12700" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
        <a:ln w="19050" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
  <a:objectDefaults/>
  <a:extraClrSchemeLst/>
</a:theme>
"@
}

function SlideLayoutXml {
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="blank" preserve="1">
  <p:cSld name="Blank">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>
"@
}

function SlideMasterXml {
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg><p:bgPr><a:solidFill><a:srgbClr val="F8FAFC"/></a:solidFill><a:effectLst/></p:bgPr></p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst>
  <p:txStyles>
    <p:titleStyle/><p:bodyStyle/><p:otherStyle/>
  </p:txStyles>
</p:sldMaster>
"@
}

function PresentationXml {
  $ids = New-Object System.Collections.Generic.List[string]
  for ($i = 1; $i -le $script:Slides.Count; $i++) {
    $ids.Add("<p:sldId id=""$([int](255 + $i))"" r:id=""rId$([int](1 + $i))""/>")
  }
  $cx = Emu $SlideW; $cy = Emu $SlideH
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>
  <p:sldIdLst>$($ids -join '')</p:sldIdLst>
  <p:sldSz cx="$cx" cy="$cy" type="wide"/>
  <p:notesSz cx="6858000" cy="9144000"/>
  <p:defaultTextStyle/>
</p:presentation>
"@
}

function PresentationRelsXml {
  $rels = New-Object System.Collections.Generic.List[string]
  $rels.Add('<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>')
  for ($i = 1; $i -le $script:Slides.Count; $i++) {
    $rels.Add("<Relationship Id=""rId$([int](1 + $i))"" Type=""http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"" Target=""slides/slide$i.xml""/>")
  }
  return "<?xml version=""1.0"" encoding=""UTF-8"" standalone=""yes""?><Relationships xmlns=""http://schemas.openxmlformats.org/package/2006/relationships"">$($rels -join '')</Relationships>"
}

function ContentTypesXml {
  $overrides = New-Object System.Collections.Generic.List[string]
  $overrides.Add('<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>')
  $overrides.Add('<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>')
  $overrides.Add('<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>')
  $overrides.Add('<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>')
  for ($i = 1; $i -le $script:Slides.Count; $i++) {
    $overrides.Add("<Override PartName=""/ppt/slides/slide$i.xml"" ContentType=""application/vnd.openxmlformats-officedocument.presentationml.slide+xml""/>")
  }
  $overrides.Add('<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>')
  $overrides.Add('<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>')
  return "<?xml version=""1.0"" encoding=""UTF-8"" standalone=""yes""?><Types xmlns=""http://schemas.openxmlformats.org/package/2006/content-types""><Default Extension=""rels"" ContentType=""application/vnd.openxmlformats-package.relationships+xml""/><Default Extension=""xml"" ContentType=""application/xml""/>$($overrides -join '')</Types>"
}

function RootRelsXml {
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>
"@
}

function CoreXml {
  $now = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Scalable Backend Development for High-Performance E-Commerce Website Platform</dc:title>
  <dc:subject>MCA Final Project Presentation</dc:subject>
  <dc:creator>Parag Gulati</dc:creator>
  <cp:lastModifiedBy>Codex</cp:lastModifiedBy>
  <dcterms:created xsi:type="dcterms:W3CDTF">$now</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">$now</dcterms:modified>
</cp:coreProperties>
"@
}

function AppXml {
  return @"
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>Codex OpenXML Generator</Application>
  <PresentationFormat>On-screen Show (16:9)</PresentationFormat>
  <Slides>$($script:Slides.Count)</Slides>
  <Company>MCA Final Project</Company>
</Properties>
"@
}

function Build-Pptx {
  Reset-Stage
  Write-Utf8 (Join-Path $StageDir "[Content_Types].xml") (ContentTypesXml)
  Write-Utf8 (Join-Path $StageDir "_rels\.rels") (RootRelsXml)
  Write-Utf8 (Join-Path $StageDir "docProps\core.xml") (CoreXml)
  Write-Utf8 (Join-Path $StageDir "docProps\app.xml") (AppXml)
  Write-Utf8 (Join-Path $StageDir "ppt\presentation.xml") (PresentationXml)
  Write-Utf8 (Join-Path $StageDir "ppt\_rels\presentation.xml.rels") (PresentationRelsXml)
  Write-Utf8 (Join-Path $StageDir "ppt\theme\theme1.xml") (ThemeXml)
  Write-Utf8 (Join-Path $StageDir "ppt\slideMasters\slideMaster1.xml") (SlideMasterXml)
  Write-Utf8 (Join-Path $StageDir "ppt\slideMasters\_rels\slideMaster1.xml.rels") "<?xml version=""1.0"" encoding=""UTF-8"" standalone=""yes""?><Relationships xmlns=""http://schemas.openxmlformats.org/package/2006/relationships""><Relationship Id=""rId1"" Type=""http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout"" Target=""../slideLayouts/slideLayout1.xml""/><Relationship Id=""rId2"" Type=""http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"" Target=""../theme/theme1.xml""/></Relationships>"
  Write-Utf8 (Join-Path $StageDir "ppt\slideLayouts\slideLayout1.xml") (SlideLayoutXml)
  Write-Utf8 (Join-Path $StageDir "ppt\slideLayouts\_rels\slideLayout1.xml.rels") "<?xml version=""1.0"" encoding=""UTF-8"" standalone=""yes""?><Relationships xmlns=""http://schemas.openxmlformats.org/package/2006/relationships""><Relationship Id=""rId1"" Type=""http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster"" Target=""../slideMasters/slideMaster1.xml""/></Relationships>"
  for ($i = 1; $i -le $script:Slides.Count; $i++) {
    Write-Utf8 (Join-Path $StageDir "ppt\slides\slide$i.xml") $script:Slides[$i - 1]
    Write-Utf8 (Join-Path $StageDir "ppt\slides\_rels\slide$i.xml.rels") "<?xml version=""1.0"" encoding=""UTF-8"" standalone=""yes""?><Relationships xmlns=""http://schemas.openxmlformats.org/package/2006/relationships""><Relationship Id=""rId1"" Type=""http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout"" Target=""../slideLayouts/slideLayout1.xml""/></Relationships>"
  }

  if (Test-Path $DeckPath) {
    Remove-Item -LiteralPath $DeckPath -Force
  }
  Assert-InWorkspace $DeckPath
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  Add-Type -AssemblyName System.IO.Compression
  $zip = [System.IO.Compression.ZipFile]::Open($DeckPath, [System.IO.Compression.ZipArchiveMode]::Create)
  try {
    Get-ChildItem -LiteralPath $StageDir -Recurse -File | ForEach-Object {
      $relative = $_.FullName.Substring($StageDir.Length + 1).Replace("\", "/")
      [System.IO.Compression.ZipFileExtensions]::CreateEntryFromFile(
        $zip,
        $_.FullName,
        $relative,
        [System.IO.Compression.CompressionLevel]::Optimal
      ) | Out-Null
    }
  } finally {
    $zip.Dispose()
  }
}

function Write-Readme {
  $content = @'
# MCA Final Project Presentation

Generated file:

- `MCA_Final_Project_Presentation_Parag_Gulati.pptx`

## Slides Created

1. Title Slide
2. Project Introduction
3. Problem Statement
4. Objectives
5. Scope of Project
6. Technology Stack
7. System Architecture
8. Modules and Services Overview
9. Database Overview
10. API and Backend Flow
11. Frontend Overview
12. Step-by-Step User Guide: Access Frontend
13. Client/User Journey: Authentication and Discovery
14. Client/User Journey: Cart, Checkout and Payment
15. Client/User Journey: Profile and Notifications
16. Seller Dashboard Journey
17. Session Analytics Dashboard Journey
18. Superadmin Journey
19. Deployment and Local Run Overview
20. Testing and Validation
21. Challenges Faced
22. Limitations and Assumptions
23. Future Enhancements
24. Conclusion
25. Thank You / Q&A

## Screenshot Placeholders To Update

Headless Chrome and Edge both exited with code 13 in this environment, so screenshots could not be captured safely. The PPTX contains clear placeholders for:

- User App Home / Product Listing
- User App Login or Signup Page
- Search / Category Filter
- Product Detail
- Cart Page
- Checkout Page
- Profile / Orders / Notification Preferences
- Seller Product Manager / Orders / Analytics
- Analytics Overview / Funnel / Heatmap
- Superadmin Users / Payments / Audit Logs

Recommended local URLs after `docker compose up -d --build`:

- User App: `http://localhost:3000`
- Seller Dashboard: `http://localhost:3001`
- Session Analytics Dashboard: `http://localhost:3002`
- Superadmin Panel: `http://localhost:3003`
- API Gateway health: `http://localhost:8080/health/ready`

## Assumptions and Evidence

- Student name is assumed as `Parag Gulati` from the provided PDF/DOCX file names.
- The presentation content is based on the repository, `FINAL_PROJECT_REPORT.md`, `api/master-api.json`, frontend route files, database docs, Docker/Kubernetes files, CI workflow, and runbooks.
- No fake credentials, fake screenshots, fake production deployment claims, or unsupported modules were added.
- Buyer seed credentials were not found in `LOCAL_ACCESS_GUIDE.md`.
- Seller, session analytics admin, and superadmin demo credentials are documented in `LOCAL_ACCESS_GUIDE.md`; they are local/demo only and are not repeated as production credentials.
- Payment providers are disabled in local defaults, so the payment flow is described as implemented/configuration-dependent.
- Superadmin search is included only as a placeholder route because the frontend route is a protected placeholder module.
- Dedicated browser E2E automation was not found; unit/component/service tests and CI checks are documented.

## Suggested Final Update Before Submission

1. Run the platform locally using Docker Compose.
2. Capture the placeholder screenshots at 16:9 or full HD resolution.
3. Replace placeholder boxes in the PPTX with real screenshots.
4. Add college name, guide name, roll number, and submission date if required by your college format.
'@
  Write-Utf8 $ReadmePath $content
}

Build-Pptx
Write-Readme

Write-Host "Created: $DeckPath"
Write-Host "Created: $ReadmePath"
