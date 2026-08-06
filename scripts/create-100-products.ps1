# create-100-products.ps1
# Run in PowerShell:
# $env:SELLER_ACCESS_TOKEN="YOUR_FRESH_ACCESS_TOKEN"
# powershell -ExecutionPolicy Bypass -File .\create-100-products.ps1

$ApiUrl = "http://localhost:8080/api/v1/seller/products"

$AccessToken = $env:SELLER_ACCESS_TOKEN
if ([string]::IsNullOrWhiteSpace($AccessToken)) {
    $AccessToken = Read-Host "Paste fresh seller access token"
}

$RunId = Get-Date -Format "yyyyMMddHHmmss"

$Headers = @{
    "Accept"        = "application/json"
    "Content-Type"  = "application/json"
    "Authorization" = "Bearer $AccessToken"
}

function New-Slug {
    param([string]$Text)

    return ($Text.ToLower() -replace "[^a-z0-9]+", "-" -replace "^-|-$", "")
}

function Test-ImageUrl {
    param([string]$Url)

    try {
        $response = Invoke-WebRequest -Uri $Url -Method Head -TimeoutSec 20 -UseBasicParsing
        if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
            return $true
        }
        return $false
    }
    catch {
        try {
            $response = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 20 -UseBasicParsing
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
                return $true
            }
            return $false
        }
        catch {
            return $false
        }
    }
}

# -------------------------------------------------------
# Real Unsplash Image Pools
# -------------------------------------------------------

$CategoryImageIndex = @{}

$FallbackImages = @(
    "https://images.unsplash.com/photo-1667838728002-80f82dfae25e?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1524758631624-e2822e304c36?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=1200&q=80"
)

$CategoryImagePools = @{
    "cat_mens_shirts" = @(
        "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1598033129183-c4f50c736f10?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1602810318383-e386cc2a3ccf?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1489987707025-afc232f7ea0f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1589310243389-96a5483213a8?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1612423284934-2850a4ea6b0f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1581655353564-df123a1eb820?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1596755094514-f87e34085b2c?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1603252109303-2751441dd157?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1554568218-0f1715e72254?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1618354691373-d851c5c3a990?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1621072156002-e2fccdc0b176?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1594938298603-c8148c4dae35?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1503341504253-dff4815485f1?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1523381210434-271e8be1f52b?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1611312449408-fcece27cdbb7?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1620012253295-c15cc3e65df4?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1618354691792-d1d42acfd860?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1618354691229-88d47f285158?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1618354691438-25bc04584c23?auto=format&fit=crop&w=1200&q=80"
    )

    "cat_womens_kurtas" = @(
        "https://images.unsplash.com/photo-1583391733956-6c78276477e2?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1610030469983-98e550d6193c?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1585487000160-6ebcfceb0d03?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1617019114583-affb34d1b3cd?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1594633312681-425c7b97ccd1?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1503342217505-b0a15ec3261c?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1515886657613-9f3515b0c78f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1496747611176-843222e1e57c?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1529139574466-a303027c1d8b?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1515372039744-b8f02a3ae446?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1550614000-4895a10e1bfd?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1485968579580-b6d095142e6e?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1539008835657-9e8e9680c956?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1524504388940-b1c1722653e1?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1524250502761-1ac6f2e30d43?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1509631179647-0177331693ae?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1520975954732-35dd22299614?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1545911825-6bfa5b0c34a9?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1558769132-cb1aea458c5e?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1483985988355-763728e1935b?auto=format&fit=crop&w=1200&q=80"
    )

    "cat_sneakers" = @(
        "https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1549298916-b41d501d3772?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1460353581641-37baddab0fa2?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1491553895911-0055eca6402d?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1608231387042-66d1773070a5?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600185365483-26d7a4cc7519?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1525966222134-fcfa99b8ae77?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1605348532760-6753d2c43329?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1543508282-6319a3e2621f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1514989940723-e8e51635b782?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1515955656352-a1fa3ffcd111?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1560769629-975ec94e6a86?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1562183241-b937e95585b6?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1579338559194-a162d19bf842?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1584735175315-9d5df23860e6?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1595341888016-a392ef81b7de?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1597045566677-8cf032ed6634?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1617606002806-94e279c22567?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1626947346165-4c2288dadc2a?auto=format&fit=crop&w=1200&q=80"
    )

    "cat_mobile_accessories" = @(
        "https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1586953208448-b95a79798f07?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1585060544812-6b45742d762f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1598327105666-5b89351aff97?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1601784551446-20c9e07cdbdb?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1580910051074-3eb694886505?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1616410011236-7a42121dd981?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1603539444875-76e7684265f6?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1556656793-08538906a9f8?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1567581935884-3349723552ca?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1561154464-82e9adf32764?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1523206489230-c012c64b2b48?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1574944985070-8f3ebc6b79d2?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1556656798-08a1a4a30f7b?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1585155770447-2f66e2a397b5?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600080972464-8e5f35f63d08?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1583863788434-e58a36330cf0?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1616348436168-de43ad0db179?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1535157412991-2ef801c1748b?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1512054502232-10a0a035d672?auto=format&fit=crop&w=1200&q=80"
    )

    "cat_home_decor" = @(
        "https://images.unsplash.com/photo-1513519245088-0e12902e5a38?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1513694203232-719a280e022f?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1524758631624-e2822e304c36?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1484101403633-562f891dc89a?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1519710164239-da123dc03ef4?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1512918728675-ed5a9ecdebfd?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1540932239986-30128078f3c5?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1533090481720-856c6e3c1fdc?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1618220179428-22790b461013?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1616486338812-3dadae4b4ace?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1616137466211-f939a420be84?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1554995207-c18c203602cb?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80",
        "https://images.unsplash.com/photo-1600210492486-724fe5c67fb0?auto=format&fit=crop&w=1200&q=80"
    )
}

function Get-ProductImages {
    param([string]$CategoryId)

    if (-not $CategoryImagePools.ContainsKey($CategoryId)) {
        throw "No image pool found for category_id: $CategoryId"
    }

    if (-not $CategoryImageIndex.ContainsKey($CategoryId)) {
        $CategoryImageIndex[$CategoryId] = 0
    }

    $pool = $CategoryImagePools[$CategoryId]

    for ($attempt = 0; $attempt -lt $pool.Count; $attempt++) {
        $index = [int]$CategoryImageIndex[$CategoryId]

        $frontImage = $pool[$index % $pool.Count]
        $backImage  = $pool[($index + 1) % $pool.Count]

        $CategoryImageIndex[$CategoryId] = $index + 1

        $frontOk = Test-ImageUrl $frontImage
        $backOk  = Test-ImageUrl $backImage

        if ($frontOk -and $backOk) {
            return @($frontImage, $backImage)
        }

        Write-Host "Image pair failed, trying next pair..." -ForegroundColor Yellow
    }

    Write-Host "Trying fallback images..." -ForegroundColor Yellow

    for ($i = 0; $i -lt $FallbackImages.Count - 1; $i++) {
        $frontImage = $FallbackImages[$i]
        $backImage  = $FallbackImages[$i + 1]

        $frontOk = Test-ImageUrl $frontImage
        $backOk  = Test-ImageUrl $backImage

        if ($frontOk -and $backOk) {
            return @($frontImage, $backImage)
        }
    }

    throw "No working image pair found for category_id: $CategoryId"
}

# -------------------------------------------------------
# Product Data - 100 Products
# 20 Men's Shirts
# 20 Women's Kurtas
# 20 Sneakers
# 20 Mobile Accessories
# 20 Home Decor
# -------------------------------------------------------

$Products = @()

# -----------------------------
# Men's Shirts - 20 products
# -----------------------------
$menShirts = @(
    @("Classic Oxford Shirt", "Acme Wear", "cotton", "regular", "blue", "M", 129900),
    @("Slim Fit Formal Shirt", "Urban Thread", "cotton blend", "slim", "white", "L", 149900),
    @("Checked Casual Shirt", "Denim Street", "cotton", "regular", "red", "M", 119900),
    @("Linen Summer Shirt", "North Club", "linen", "relaxed", "beige", "L", 179900),
    @("Denim Button Down Shirt", "Rugged Co", "denim", "regular", "blue", "XL", 189900),
    @("Solid Black Shirt", "Metro Style", "cotton", "slim", "black", "M", 139900),
    @("Printed Vacation Shirt", "Coastline", "rayon", "relaxed", "green", "L", 159900),
    @("Mandarin Collar Shirt", "Indigo Lane", "cotton", "regular", "navy", "M", 169900),
    @("Flannel Winter Shirt", "HillWear", "flannel", "regular", "brown", "XL", 199900),
    @("Office Stripe Shirt", "WorkMode", "cotton blend", "slim", "sky blue", "L", 154900),
    @("Premium White Shirt", "Executive Fit", "cotton", "regular", "white", "M", 209900),
    @("Olive Casual Shirt", "StreetForm", "cotton", "regular", "olive", "L", 134900),
    @("Navy Formal Shirt", "Boardroom", "cotton blend", "slim", "navy", "M", 159900),
    @("Half Sleeve Shirt", "DailyBasics", "cotton", "regular", "yellow", "M", 99900),
    @("Textured Party Shirt", "NightOut", "poly cotton", "slim", "maroon", "L", 179900),
    @("Micro Print Shirt", "Urban Thread", "cotton", "regular", "grey", "M", 149900),
    @("Relaxed Camp Shirt", "Coastline", "viscose", "relaxed", "cream", "L", 169900),
    @("Formal Grey Shirt", "Executive Fit", "cotton blend", "regular", "grey", "XL", 164900),
    @("Blue Chambray Shirt", "Rugged Co", "chambray", "regular", "blue", "L", 189900),
    @("Minimal Beige Shirt", "North Club", "linen blend", "relaxed", "beige", "M", 174900)
)

foreach ($item in $menShirts) {
    $title = "Men's $($item[0])"

    $Products += @{
        title = $title
        description = "$title made with $($item[2]) fabric, designed for comfortable daily wear."
        brand = $item[1]
        category_id = "cat_mens_shirts"
        attributes = @{
            material = $item[2]
            fit = $item[3]
            color_family = $item[4]
            gender = "men"
        }
        variants = @(
            @{
                sku = "MSHIRT-$RunId-$((New-Slug $item[0]).ToUpper())"
                attributes = @{
                    color = $item[4]
                    size = $item[5]
                }
                price = @{
                    amount = [int]$item[6]
                    currency = "INR"
                }
                stock_quantity = 25
            }
        )
    }
}

# -----------------------------
# Women's Kurtas - 20 products
# -----------------------------
$kurtas = @(
    @("Floral Cotton Kurta", "Rangriti", "cotton", "regular", "pink", "M", 119900),
    @("Straight Rayon Kurta", "EthnicAura", "rayon", "straight", "blue", "L", 139900),
    @("Printed A-Line Kurta", "Nayra", "viscose", "a-line", "green", "M", 149900),
    @("Embroidered Festive Kurta", "Saanvi", "cotton silk", "regular", "maroon", "L", 229900),
    @("Daily Wear Kurta", "Rangriti", "cotton", "regular", "yellow", "M", 99900),
    @("Anarkali Kurta", "EthnicAura", "rayon", "anarkali", "purple", "L", 199900),
    @("Chikankari Style Kurta", "Nayra", "cotton", "straight", "white", "M", 179900),
    @("Block Print Kurta", "Jaipur Loom", "cotton", "straight", "indigo", "L", 159900),
    @("Office Wear Kurta", "WorkEthnic", "cotton blend", "straight", "grey", "M", 129900),
    @("Pastel Kurta", "Saanvi", "rayon", "regular", "peach", "L", 144900),
    @("Long Kurta", "EthnicAura", "cotton", "straight", "black", "XL", 169900),
    @("Short Kurta", "DailyBasics", "cotton", "regular", "red", "M", 89900),
    @("Festive Printed Kurta", "Jaipur Loom", "cotton silk", "a-line", "orange", "L", 219900),
    @("Minimal Solid Kurta", "WorkEthnic", "rayon", "straight", "navy", "M", 124900),
    @("Thread Work Kurta", "Saanvi", "viscose", "regular", "cream", "L", 189900),
    @("Geometric Print Kurta", "Nayra", "cotton", "straight", "teal", "M", 139900),
    @("Casual Slub Kurta", "Rangriti", "slub cotton", "regular", "olive", "L", 134900),
    @("Designer Neck Kurta", "EthnicAura", "rayon", "a-line", "magenta", "M", 169900),
    @("Summer Kurta", "DailyBasics", "cotton", "regular", "sky blue", "L", 109900),
    @("Traditional Kurta", "Jaipur Loom", "cotton", "straight", "mustard", "M", 159900)
)

foreach ($item in $kurtas) {
    $title = "Women's $($item[0])"

    $Products += @{
        title = $title
        description = "$title with elegant design and breathable $($item[2]) fabric."
        brand = $item[1]
        category_id = "cat_womens_kurtas"
        attributes = @{
            material = $item[2]
            fit = $item[3]
            color_family = $item[4]
            gender = "women"
        }
        variants = @(
            @{
                sku = "WKURTA-$RunId-$((New-Slug $item[0]).ToUpper())"
                attributes = @{
                    color = $item[4]
                    size = $item[5]
                }
                price = @{
                    amount = [int]$item[6]
                    currency = "INR"
                }
                stock_quantity = 30
            }
        )
    }
}

# -----------------------------
# Sneakers - 20 products
# -----------------------------
$sneakers = @(
    @("Classic White Sneakers", "StrideX", "synthetic leather", "low-top", "white", "9", 249900),
    @("Running Mesh Sneakers", "RunPro", "mesh", "sport", "black", "8", 299900),
    @("Street High Top Sneakers", "UrbanKicks", "canvas", "high-top", "red", "9", 279900),
    @("Casual Grey Sneakers", "WalkMate", "knit", "low-top", "grey", "10", 219900),
    @("Chunky Sole Sneakers", "StreetForm", "synthetic", "chunky", "beige", "8", 349900),
    @("Navy Everyday Sneakers", "StrideX", "canvas", "low-top", "navy", "9", 229900),
    @("Training Sneakers", "RunPro", "mesh", "training", "blue", "10", 329900),
    @("Slip-On Sneakers", "WalkMate", "knit", "slip-on", "black", "8", 199900),
    @("Retro Court Sneakers", "UrbanKicks", "synthetic leather", "court", "white", "9", 319900),
    @("Outdoor Grip Sneakers", "TrailGo", "mesh", "outdoor", "olive", "10", 359900),
    @("Minimal Beige Sneakers", "StrideX", "canvas", "low-top", "beige", "8", 239900),
    @("Black Street Sneakers", "StreetForm", "synthetic", "street", "black", "9", 269900),
    @("Lightweight Jogger Sneakers", "RunPro", "mesh", "running", "grey", "8", 289900),
    @("Canvas Daily Sneakers", "WalkMate", "canvas", "casual", "green", "9", 189900),
    @("Premium Leather Sneakers", "UrbanKicks", "leather", "low-top", "brown", "10", 449900),
    @("Sporty Red Sneakers", "RunPro", "mesh", "sport", "red", "8", 299900),
    @("Skate Style Sneakers", "StreetForm", "suede", "skate", "black", "9", 259900),
    @("Breathable Knit Sneakers", "WalkMate", "knit", "low-top", "blue", "10", 249900),
    @("Monochrome Sneakers", "StrideX", "synthetic leather", "low-top", "black", "9", 279900),
    @("Trail Runner Sneakers", "TrailGo", "mesh", "trail", "orange", "10", 389900)
)

foreach ($item in $sneakers) {
    $title = $item[0]

    $Products += @{
        title = $title
        description = "$title with durable $($item[2]) upper and comfortable sole for daily use."
        brand = $item[1]
        category_id = "cat_sneakers"
        attributes = @{
            material = $item[2]
            style = $item[3]
            color_family = $item[4]
            gender = "unisex"
        }
        variants = @(
            @{
                sku = "SNKR-$RunId-$((New-Slug $item[0]).ToUpper())"
                attributes = @{
                    color = $item[4]
                    size = $item[5]
                }
                price = @{
                    amount = [int]$item[6]
                    currency = "INR"
                }
                stock_quantity = 20
            }
        )
    }
}

# -----------------------------
# Mobile Accessories - 20 products
# -----------------------------
$mobileAccessories = @(
    @("Fast Charging Cable", "VoltEdge", "nylon braided", "type-c", "black", 49900),
    @("20W USB-C Charger", "ChargeMax", "plastic", "charger", "white", 89900),
    @("Magnetic Phone Holder", "GripGo", "aluminium", "holder", "black", 69900),
    @("Clear Phone Case", "CaseCraft", "TPU", "case", "transparent", 39900),
    @("Tempered Glass Protector", "ShieldPro", "glass", "screen protector", "clear", 29900),
    @("Wireless Charging Pad", "VoltEdge", "plastic", "wireless charger", "black", 129900),
    @("Power Bank 10000mAh", "ChargeMax", "ABS plastic", "power bank", "blue", 189900),
    @("Bluetooth Selfie Stick", "SnapMate", "aluminium", "selfie stick", "black", 79900),
    @("Phone Ring Holder", "GripGo", "metal", "holder", "silver", 24900),
    @("Camera Lens Protector", "ShieldPro", "glass", "lens protector", "clear", 34900),
    @("Dual Port Car Charger", "ChargeMax", "plastic", "car charger", "black", 99900),
    @("Silicone Phone Case", "CaseCraft", "silicone", "case", "red", 49900),
    @("Earphone Cable Organizer", "DailyTech", "silicone", "organizer", "grey", 19900),
    @("Mobile Cleaning Kit", "DailyTech", "microfiber", "cleaning kit", "white", 29900),
    @("Adjustable Desk Stand", "GripGo", "aluminium", "stand", "silver", 89900),
    @("Fast Charging Adapter", "VoltEdge", "plastic", "adapter", "white", 79900),
    @("Braided Lightning Cable", "VoltEdge", "nylon braided", "lightning cable", "black", 59900),
    @("Anti Glare Screen Guard", "ShieldPro", "glass", "screen protector", "clear", 39900),
    @("Mini Power Bank", "ChargeMax", "ABS plastic", "power bank", "black", 149900),
    @("Universal Mobile Stand", "DailyTech", "plastic", "stand", "blue", 34900)
)

foreach ($item in $mobileAccessories) {
    $title = $item[0]

    $Products += @{
        title = $title
        description = "$title designed for reliable mobile usage and everyday convenience."
        brand = $item[1]
        category_id = "cat_mobile_accessories"
        attributes = @{
            material = $item[2]
            accessory_type = $item[3]
            color_family = $item[4]
        }
        variants = @(
            @{
                sku = "MACC-$RunId-$((New-Slug $item[0]).ToUpper())"
                attributes = @{
                    color = $item[4]
                    type = $item[3]
                }
                price = @{
                    amount = [int]$item[5]
                    currency = "INR"
                }
                stock_quantity = 50
            }
        )
    }
}

# -----------------------------
# Home Decor - 20 products
# -----------------------------
$homeDecor = @(
    @("Ceramic Flower Vase", "HomeAura", "ceramic", "vase", "white", 99900),
    @("Wall Art Frame", "DecorNest", "wood", "wall art", "brown", 129900),
    @("LED String Lights", "GlowHome", "copper wire", "lights", "warm white", 69900),
    @("Decorative Cushion Cover", "SoftNest", "cotton", "cushion cover", "blue", 39900),
    @("Table Candle Holder", "HomeAura", "metal", "candle holder", "gold", 59900),
    @("Wooden Wall Shelf", "DecorNest", "engineered wood", "shelf", "brown", 149900),
    @("Artificial Plant Pot", "GreenDecor", "plastic", "plant", "green", 79900),
    @("Macrame Wall Hanging", "CraftLane", "cotton rope", "wall hanging", "cream", 89900),
    @("Scented Candle Set", "GlowHome", "wax", "candle", "lavender", 49900),
    @("Decorative Mirror", "DecorNest", "glass", "mirror", "silver", 219900),
    @("Photo Frame Set", "HomeAura", "wood", "photo frame", "black", 99900),
    @("Cotton Table Runner", "SoftNest", "cotton", "table runner", "beige", 69900),
    @("Metal Wall Clock", "DecorNest", "metal", "clock", "black", 159900),
    @("Boho Cushion Cover", "CraftLane", "cotton", "cushion cover", "multi", 44900),
    @("Hanging Planter", "GreenDecor", "metal", "planter", "white", 89900),
    @("Decorative Lantern", "GlowHome", "metal", "lantern", "gold", 119900),
    @("Abstract Wall Painting", "DecorNest", "canvas", "painting", "multi", 179900),
    @("Marble Look Tray", "HomeAura", "resin", "tray", "white", 99900),
    @("Mini Desk Plant", "GreenDecor", "plastic", "plant", "green", 34900),
    @("Luxury Candle Jar", "GlowHome", "glass", "candle", "amber", 79900)
)

foreach ($item in $homeDecor) {
    $title = $item[0]

    $Products += @{
        title = $title
        description = "$title crafted with $($item[2]) material to enhance modern home interiors."
        brand = $item[1]
        category_id = "cat_home_decor"
        attributes = @{
            material = $item[2]
            decor_type = $item[3]
            color_family = $item[4]
        }
        variants = @(
            @{
                sku = "HDECOR-$RunId-$((New-Slug $item[0]).ToUpper())"
                attributes = @{
                    color = $item[4]
                    type = $item[3]
                }
                price = @{
                    amount = [int]$item[5]
                    currency = "INR"
                }
                stock_quantity = 35
            }
        )
    }
}

# -------------------------------------------------------
# API Execution
# -------------------------------------------------------

Write-Host "====================================="
Write-Host "Total products prepared: $($Products.Count)"
Write-Host "API URL: $ApiUrl"
Write-Host "====================================="

$SuccessCount = 0
$FailCount = 0
$FailedProducts = @()

foreach ($product in $Products) {
    Write-Host "`nPreparing product: $($product["title"])"

    try {
        $images = Get-ProductImages -CategoryId $product["category_id"]
        $product["images"] = @($images[0], $images[1])

        Write-Host "Image 1: $($images[0])"
        Write-Host "Image 2: $($images[1])"
    }
    catch {
        $FailCount++

        Write-Host "Image failed for product: $($product["title"])" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red

        $FailedProducts += @{
            title = $product["title"]
            category_id = $product["category_id"]
            reason = $_.Exception.Message
        }

        continue
    }

    $JsonBody = $product | ConvertTo-Json -Depth 20

    try {
        Write-Host "Creating product: $($product["title"])"

        $response = Invoke-RestMethod `
            -Uri $ApiUrl `
            -Method Post `
            -Headers $Headers `
            -Body $JsonBody `
            -TimeoutSec 60

        $SuccessCount++
        Write-Host "Created successfully: $($product["title"])" -ForegroundColor Green
    }
    catch {
        $FailCount++

        $errorMessage = $_.Exception.Message
        if ($_.ErrorDetails.Message) {
            $errorMessage = $_.ErrorDetails.Message
        }

        Write-Host "Failed: $($product["title"])" -ForegroundColor Red
        Write-Host $errorMessage -ForegroundColor Red

        $FailedProducts += @{
            title = $product["title"]
            category_id = $product["category_id"]
            reason = $errorMessage
            body = $product
        }
    }
}

Write-Host "`n====================================="
Write-Host "Product Creation Completed"
Write-Host "Success: $SuccessCount"
Write-Host "Failed : $FailCount"
Write-Host "====================================="

if ($FailedProducts.Count -gt 0) {
    $failedFile = ".\failed-products-$RunId.json"
    $FailedProducts | ConvertTo-Json -Depth 20 | Out-File $failedFile
    Write-Host "Failed products saved to: $failedFile" -ForegroundColor Yellow
}