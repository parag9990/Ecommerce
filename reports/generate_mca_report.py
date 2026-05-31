#!/usr/bin/env python3
"""Generate the MCA Major Project Report PDF without external PDF packages.

The execution environment for this project does not include LaTeX, Pandoc,
ReportLab, or Poppler tools. This script writes a simple, standards-compliant
PDF directly using built-in Python only. The layout follows the uploaded MCA
Major Project guideline as closely as possible from the extracted text:

- A4 page size
- Times family fonts
- 12 pt body text
- 1.5 line spacing
- 1.25 inch left and right margins
- bottom-right "Page | n" numbering
- required report order from Title Page to Bibliography
"""

from __future__ import annotations

import math
import textwrap
from dataclasses import dataclass, field
from pathlib import Path
from typing import Callable


PROJECT_TITLE = "Scalable Backend Development for High-Performance E-Commerce Website Platform"
SHORT_TITLE = "Scalable E-Commerce Platform"
OUTPUT_PATH = Path("reports/MCA_Major_Project_Report.pdf")

PAGE_W = 595.28
PAGE_H = 841.89
MARGIN_L = 90.0
MARGIN_R = 90.0
MARGIN_T = 90.0
MARGIN_B = 78.0
CONTENT_W = PAGE_W - MARGIN_L - MARGIN_R

FONT_REG = "F1"
FONT_BOLD = "F2"
FONT_ITALIC = "F3"
FONT_MONO = "F4"


def clean(text: object) -> str:
    """Keep output ASCII and PDF-safe."""
    s = str(text)
    replacements = {
        "\u2013": "-",
        "\u2014": "-",
        "\u2018": "'",
        "\u2019": "'",
        "\u201c": '"',
        "\u201d": '"',
        "\u2022": "-",
        "\u00a0": " ",
        "\u00b1": "+/-",
        "\u2192": "->",
        "\u2265": ">=",
        "\u2264": "<=",
    }
    for src, dst in replacements.items():
        s = s.replace(src, dst)
    return s.encode("ascii", "replace").decode("ascii").replace("?", "")


def pdf_escape(text: str) -> str:
    text = clean(text)
    return text.replace("\\", "\\\\").replace("(", "\\(").replace(")", "\\)")


@dataclass
class Page:
    number: int
    ops: list[str] = field(default_factory=list)

    def add(self, op: str) -> None:
        self.ops.append(op)


class ReportDoc:
    def __init__(self, toc_pages: dict[str, int] | None = None):
        self.pages: list[Page] = []
        self.page: Page | None = None
        self.y = PAGE_H - MARGIN_T
        self.toc_pages = toc_pages or {}
        self.toc_entries: list[tuple[str, int, int]] = []
        self.figure_no = 0
        self.add_page(show_header=False)

    def add_page(self, show_header: bool = True) -> None:
        page = Page(len(self.pages) + 1)
        self.pages.append(page)
        self.page = page
        self.y = PAGE_H - MARGIN_T
        if show_header:
            self.header()
        self.footer()

    def header(self) -> None:
        assert self.page is not None
        y = PAGE_H - 42
        self.text(MARGIN_L, y, "MCA Major Project Report", 9, FONT_ITALIC)
        self.text(PAGE_W - MARGIN_R - 170, y, "Centre for Distance & Online Education", 9, FONT_ITALIC)
        self.line(MARGIN_L, y - 10, PAGE_W - MARGIN_R, y - 10, 0.35, (0.55, 0.55, 0.55))

    def footer(self) -> None:
        assert self.page is not None
        self.text(PAGE_W - MARGIN_R - 38, 42, f"Page | {self.page.number}", 10, FONT_REG)

    def ensure(self, need: float, header: bool = True) -> None:
        if self.y - need < MARGIN_B:
            self.add_page(show_header=header)

    def mark(self, title: str, level: int = 0) -> None:
        assert self.page is not None
        self.toc_entries.append((title, level, self.page.number))

    def text(self, x: float, y: float, text: str, size: float = 12, font: str = FONT_REG,
             color: tuple[float, float, float] = (0, 0, 0)) -> None:
        assert self.page is not None
        r, g, b = color
        self.page.add(f"{r:.3f} {g:.3f} {b:.3f} rg BT /{font} {size:.2f} Tf 1 0 0 1 {x:.2f} {y:.2f} Tm ({pdf_escape(text)}) Tj ET 0 0 0 rg")

    def text_center(self, x: float, y: float, w: float, text: str, size: float = 12,
                    font: str = FONT_REG, color: tuple[float, float, float] = (0, 0, 0)) -> None:
        lines = self.wrap(text, w - 8, size)
        total = len(lines) * size * 1.05
        yy = y + total / 2 - size
        for line in lines:
            approx = len(clean(line)) * size * 0.48
            self.text(x + (w - approx) / 2, yy, line, size, font, color)
            yy -= size * 1.10

    def line(self, x1: float, y1: float, x2: float, y2: float, width: float = 0.7,
             color: tuple[float, float, float] = (0, 0, 0)) -> None:
        assert self.page is not None
        r, g, b = color
        self.page.add(f"q {width:.2f} w {r:.3f} {g:.3f} {b:.3f} RG {x1:.2f} {y1:.2f} m {x2:.2f} {y2:.2f} l S Q")

    def rect(self, x: float, y: float, w: float, h: float, stroke: tuple[float, float, float] = (0, 0, 0),
             fill: tuple[float, float, float] | None = None, width: float = 0.7) -> None:
        assert self.page is not None
        if fill is None:
            fill_ops = ""
            draw = "S"
        else:
            fill_ops = f"{fill[0]:.3f} {fill[1]:.3f} {fill[2]:.3f} rg "
            draw = "B"
        self.page.add(
            f"q {width:.2f} w {stroke[0]:.3f} {stroke[1]:.3f} {stroke[2]:.3f} RG "
            f"{fill_ops}{x:.2f} {y:.2f} {w:.2f} {h:.2f} re {draw} Q"
        )

    def ellipse(self, cx: float, cy: float, rx: float, ry: float,
                stroke: tuple[float, float, float] = (0, 0, 0),
                fill: tuple[float, float, float] | None = None, width: float = 0.7) -> None:
        # Approximate ellipse with four Bezier curves.
        k = 0.5522847498
        x0, y0 = cx - rx, cy
        cmds = [
            f"{x0:.2f} {y0:.2f} m",
            f"{x0:.2f} {cy + k*ry:.2f} {cx - k*rx:.2f} {cy + ry:.2f} {cx:.2f} {cy + ry:.2f} c",
            f"{cx + k*rx:.2f} {cy + ry:.2f} {cx + rx:.2f} {cy + k*ry:.2f} {cx + rx:.2f} {cy:.2f} c",
            f"{cx + rx:.2f} {cy - k*ry:.2f} {cx + k*rx:.2f} {cy - ry:.2f} {cx:.2f} {cy - ry:.2f} c",
            f"{cx - k*rx:.2f} {cy - ry:.2f} {x0:.2f} {cy - k*ry:.2f} {x0:.2f} {cy:.2f} c",
        ]
        fill_ops = ""
        draw = "S"
        if fill is not None:
            fill_ops = f"{fill[0]:.3f} {fill[1]:.3f} {fill[2]:.3f} rg "
            draw = "B"
        assert self.page is not None
        self.page.add(
            f"q {width:.2f} w {stroke[0]:.3f} {stroke[1]:.3f} {stroke[2]:.3f} RG {fill_ops}"
            + " ".join(cmds)
            + f" {draw} Q"
        )

    def arrow(self, x1: float, y1: float, x2: float, y2: float, width: float = 0.8,
              color: tuple[float, float, float] = (0, 0, 0)) -> None:
        self.line(x1, y1, x2, y2, width, color)
        angle = math.atan2(y2 - y1, x2 - x1)
        size = 6
        left = angle + math.pi * 0.82
        right = angle - math.pi * 0.82
        lx = x2 + math.cos(left) * size
        ly = y2 + math.sin(left) * size
        rx = x2 + math.cos(right) * size
        ry = y2 + math.sin(right) * size
        self.line(x2, y2, lx, ly, width, color)
        self.line(x2, y2, rx, ry, width, color)

    def wrap(self, text: str, width_pts: float, size: float) -> list[str]:
        max_chars = max(18, int(width_pts / (size * 0.48)))
        return textwrap.wrap(clean(text), width=max_chars, break_long_words=False, replace_whitespace=True) or [""]

    def paragraph(self, text: str, size: float = 12, indent: float = 0, leading: float | None = None,
                  font: str = FONT_REG, gap: float = 7) -> None:
        leading = leading or size * 1.5
        for para in clean(text).split("\n"):
            para = para.strip()
            if not para:
                self.y -= leading * 0.5
                continue
            lines = self.wrap(para, CONTENT_W - indent, size)
            self.ensure(len(lines) * leading + gap)
            first = True
            for line in lines:
                self.text(MARGIN_L + (indent if first else indent), self.y, line, size, font)
                self.y -= leading
                first = False
            self.y -= gap

    def bullet(self, text: str, size: float = 12, indent: float = 18) -> None:
        lines = self.wrap(text, CONTENT_W - indent - 10, size)
        self.ensure(len(lines) * size * 1.5 + 3)
        self.text(MARGIN_L, self.y, "-", size, FONT_REG)
        self.text(MARGIN_L + indent, self.y, lines[0], size, FONT_REG)
        self.y -= size * 1.5
        for line in lines[1:]:
            self.text(MARGIN_L + indent, self.y, line, size, FONT_REG)
            self.y -= size * 1.5
        self.y -= 2

    def heading(self, text: str, level: int = 1, mark: bool = False) -> None:
        if level == 1:
            self.ensure(80)
            size = 16
            gap_before = 10
            gap_after = 12
        elif level == 2:
            self.ensure(58)
            size = 14
            gap_before = 8
            gap_after = 8
        else:
            self.ensure(44)
            size = 12.5
            gap_before = 5
            gap_after = 6
        self.y -= gap_before
        self.text(MARGIN_L, self.y, text, size, FONT_BOLD)
        if level == 1:
            self.line(MARGIN_L, self.y - 5, PAGE_W - MARGIN_R, self.y - 5, 0.55, (0.25, 0.25, 0.25))
        self.y -= size * 1.5 + gap_after
        if mark:
            self.mark(text, level - 1)

    def table(self, headers: list[str], rows: list[list[str]], widths: list[float],
              font_size: float = 9.2) -> None:
        assert abs(sum(widths) - CONTENT_W) < 2.0, "Table widths must match content width"

        def row_lines(values: list[str]) -> list[list[str]]:
            return [self.wrap(v, w - 8, font_size) for v, w in zip(values, widths)]

        def draw_row(values: list[str], y_top: float, fill: tuple[float, float, float] | None,
                     bold: bool = False) -> float:
            cell_lines = row_lines(values)
            max_lines = max(len(lines) for lines in cell_lines)
            h = max(20, max_lines * font_size * 1.28 + 10)
            x = MARGIN_L
            for idx, (v, w) in enumerate(zip(values, widths)):
                self.rect(x, y_top - h, w, h, stroke=(0.25, 0.25, 0.25), fill=fill, width=0.45)
                yy = y_top - font_size - 6
                for line in cell_lines[idx]:
                    self.text(x + 4, yy, line, font_size, FONT_BOLD if bold else FONT_REG)
                    yy -= font_size * 1.22
                x += w
            return h

        header_h = 24
        self.ensure(header_h + 20)
        h = draw_row(headers, self.y, (0.88, 0.91, 0.94), bold=True)
        self.y -= h
        for row in rows:
            lines = row_lines(row)
            need = max(20, max(len(cell) for cell in lines) * font_size * 1.28 + 10)
            if self.y - need < MARGIN_B:
                self.add_page()
                h = draw_row(headers, self.y, (0.88, 0.91, 0.94), bold=True)
                self.y -= h
            h = draw_row(row, self.y, None, bold=False)
            self.y -= h
        self.y -= 12

    def code_block(self, lines: list[str], size: float = 8.6) -> None:
        leading = size * 1.35
        h = len(lines) * leading + 14
        self.ensure(h + 10)
        self.rect(MARGIN_L, self.y - h + 5, CONTENT_W, h, stroke=(0.45, 0.45, 0.45), fill=(0.96, 0.96, 0.96), width=0.45)
        yy = self.y - 10
        for line in lines:
            self.text(MARGIN_L + 8, yy, line[:92], size, FONT_MONO)
            yy -= leading
        self.y -= h + 8

    def figure_page(self, title: str, draw_func: Callable[["ReportDoc", float, float, float, float], None],
                    explanation: str) -> None:
        self.add_page()
        self.figure_no += 1
        self.heading(f"Figure {self.figure_no}: {title}", 2)
        top = self.y
        height = 430
        self.rect(MARGIN_L, top - height, CONTENT_W, height, stroke=(0.35, 0.35, 0.35), fill=(0.99, 0.99, 0.99), width=0.5)
        draw_func(self, MARGIN_L + 12, top - 16, CONTENT_W - 24, height - 32)
        self.y = top - height - 18
        self.paragraph(f"Explanation: {explanation}", size=11.5)

    def write_toc(self) -> None:
        self.heading("Table of Contents", 1, mark=True)
        rows = []
        for title, level, page in self.toc_pages.get("entries", []):
            label = ("   " * level) + title
            rows.append([label, str(page)])
        self.table(["Section", "Page"], rows, [CONTENT_W - 55, 55], font_size=10.5)


def box(doc: ReportDoc, x: float, y_top: float, w: float, h: float, text: str,
        fill: tuple[float, float, float] = (0.91, 0.95, 0.99),
        font_size: float = 9.0, bold: bool = True) -> tuple[float, float]:
    doc.rect(x, y_top - h, w, h, stroke=(0.16, 0.24, 0.32), fill=fill, width=0.75)
    doc.text_center(x, y_top - h / 2 + font_size / 2, w, text, font_size, FONT_BOLD if bold else FONT_REG)
    return (x + w / 2, y_top - h / 2)


def draw_system_architecture(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    top = y
    c_user = box(doc, x + 5, top, 76, 36, "User App", (0.90, 0.96, 0.92))
    c_seller = box(doc, x + 95, top, 88, 36, "Seller Dashboard", (0.90, 0.96, 0.92))
    c_admin = box(doc, x + 198, top, 88, 36, "Superadmin", (0.90, 0.96, 0.92))
    c_analytics = box(doc, x + 302, top, 96, 36, "Analytics Dashboard", (0.90, 0.96, 0.92))
    c_lb = box(doc, x + 135, top - 70, 142, 34, "Load Balancer / Ingress", (0.96, 0.94, 0.88))
    c_gw = box(doc, x + 139, top - 125, 132, 38, "API Gateway", (0.90, 0.93, 0.98))
    for c in [c_user, c_seller, c_admin, c_analytics]:
        doc.arrow(c[0], c[1] - 18, c_lb[0], c_lb[1] + 17)
    doc.arrow(c_lb[0], c_lb[1] - 17, c_gw[0], c_gw[1] + 19)
    svc_y = top - 200
    services = [
        ("Auth", 0), ("User", 73), ("Product", 146), ("Cart", 219), ("Order", 292), ("Payment", 365)
    ]
    centers = []
    for name, dx in services:
        centers.append(box(doc, x + dx, svc_y, 61, 34, name, (0.92, 0.96, 1.0), 8.5))
    for c in centers:
        doc.arrow(c_gw[0], c_gw[1] - 19, c[0], c[1] + 17)
    svc2_y = top - 258
    services2 = [("Search", 0), ("CMS", 73), ("Session", 146), ("Notification", 219), ("Recommendation", 320)]
    centers2 = []
    for name, dx in services2:
        ww = 88 if name in {"Notification", "Recommendation"} else 61
        centers2.append(box(doc, x + dx, svc2_y, ww, 34, name, (0.92, 0.96, 1.0), 8.3))
    for c in centers2:
        doc.arrow(c_gw[0], c_gw[1] - 19, c[0], c[1] + 17)
    data_y = top - 345
    dbs = [("MySQL", 20), ("MongoDB", 113), ("Redis", 210), ("Typesense", 300), ("Kafka/RabbitMQ", 385)]
    for name, dx in dbs:
        ww = 86 if name == "Kafka/RabbitMQ" else 74
        c = box(doc, x + dx, data_y, ww, 38, name, (0.98, 0.94, 0.92), 8.8)
        doc.arrow(c[0], c[1] + 19, c[0], c[1] + 63)


def draw_workflow(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    steps = [
        ("Browse/Search", "User searches products"),
        ("Cart", "Add or update items"),
        ("Checkout", "Create order request"),
        ("Auth", "Verify token/session"),
        ("Inventory", "Reserve product stock"),
        ("Payment", "Create payment intent"),
        ("Webhook", "Confirm payment"),
        ("Order Paid", "Update order status"),
        ("Notify", "Send confirmation"),
    ]
    start_x = x + 8
    top = y - 18
    centers = []
    for i, (head, sub) in enumerate(steps):
        col = i % 3
        row = i // 3
        xx = start_x + col * 137
        yy = top - row * 118
        c = box(doc, xx, yy, 112, 52, f"{head}\n{sub}", (0.93, 0.96, 0.99), 8.3)
        centers.append(c)
    for i in range(len(centers) - 1):
        doc.arrow(centers[i][0] + 56 if i % 3 != 2 else centers[i][0],
                  centers[i][1],
                  centers[i + 1][0] - 56 if (i + 1) % 3 != 0 else centers[i + 1][0],
                  centers[i + 1][1])


def draw_use_case(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    left_actor = box(doc, x + 8, y - 40, 78, 36, "Buyer", (0.96, 0.94, 0.88))
    seller = box(doc, x + 8, y - 145, 78, 36, "Seller", (0.96, 0.94, 0.88))
    admin = box(doc, x + 8, y - 250, 78, 36, "Admin", (0.96, 0.94, 0.88))
    provider = box(doc, x + w - 92, y - 145, 78, 36, "Payment Provider", (0.96, 0.94, 0.88), 8.0)
    cases = [
        ("Register/Login", x + 150, y - 35),
        ("Browse and Search", x + 150, y - 90),
        ("Manage Cart", x + 150, y - 145),
        ("Checkout and Pay", x + 150, y - 200),
        ("Manage Products", x + 290, y - 90),
        ("View Analytics", x + 290, y - 145),
        ("Manage Users", x + 290, y - 200),
        ("Refund Review", x + 290, y - 255),
    ]
    centers = []
    for label, cx, cy in cases:
        doc.ellipse(cx, cy, 58, 20, stroke=(0.16, 0.24, 0.32), fill=(0.93, 0.96, 1.0), width=0.75)
        doc.text_center(cx - 58, cy + 4, 116, label, 8.2, FONT_BOLD)
        centers.append((label, cx, cy))
    for _, cx, cy in centers[:4]:
        doc.line(left_actor[0] + 39, left_actor[1], cx - 58, cy, 0.5)
    for label, cx, cy in centers:
        if label in {"Manage Products", "View Analytics"}:
            doc.line(seller[0] + 39, seller[1], cx - 58, cy, 0.5)
        if label in {"Manage Users", "Refund Review", "View Analytics"}:
            doc.line(admin[0] + 39, admin[1], cx - 58, cy, 0.5)
        if label in {"Checkout and Pay", "Refund Review"}:
            doc.line(cx + 58, cy, provider[0] - 39, provider[1], 0.5)


def draw_dfd(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    user = box(doc, x + 10, y - 75, 86, 42, "Customer / Seller / Admin", (0.96, 0.94, 0.88), 8.4)
    platform = box(doc, x + 158, y - 70, 122, 55, "E-Commerce Platform", (0.90, 0.93, 0.98), 9.5)
    payment = box(doc, x + 333, y - 50, 90, 38, "Payment Provider", (0.96, 0.94, 0.88), 8.3)
    notif = box(doc, x + 333, y - 112, 90, 38, "Email/SMS Service", (0.96, 0.94, 0.88), 8.3)
    data1 = box(doc, x + 80, y - 220, 82, 42, "MySQL Stores", (0.98, 0.94, 0.92), 8.6)
    data2 = box(doc, x + 190, y - 220, 82, 42, "Mongo Stores", (0.98, 0.94, 0.92), 8.6)
    data3 = box(doc, x + 300, y - 220, 82, 42, "Redis / Queue", (0.98, 0.94, 0.92), 8.6)
    doc.arrow(user[0] + 43, user[1], platform[0] - 61, platform[1])
    doc.arrow(platform[0] - 61, platform[1] + 12, user[0] + 43, user[1] + 12)
    doc.arrow(platform[0] + 61, platform[1] + 12, payment[0] - 45, payment[1])
    doc.arrow(payment[0] - 45, payment[1] - 10, platform[0] + 61, platform[1] - 10)
    doc.arrow(platform[0] + 61, platform[1] - 22, notif[0] - 45, notif[1])
    for c in [data1, data2, data3]:
        doc.arrow(platform[0], platform[1] - 28, c[0], c[1] + 21)


def draw_er(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    entities = {
        "AuthAccount": (x + 8, y - 12),
        "Credential": (x + 145, y - 12),
        "RefreshToken": (x + 282, y - 12),
        "RoleAssignment": (x + 8, y - 100),
        "OTPChallenge": (x + 145, y - 100),
        "User": (x + 282, y - 100),
        "SellerProfile": (x + 8, y - 188),
        "Order": (x + 145, y - 188),
        "OrderItem": (x + 282, y - 188),
        "Payment": (x + 145, y - 276),
    }
    centers = {}
    for name, (xx, yy) in entities.items():
        c = box(doc, xx, yy, 108, 48, name, (0.93, 0.96, 1.0), 8.5)
        centers[name] = c
    links = [
        ("AuthAccount", "Credential", "1:1"),
        ("AuthAccount", "RefreshToken", "1:M"),
        ("AuthAccount", "RoleAssignment", "1:M"),
        ("AuthAccount", "OTPChallenge", "1:M"),
        ("AuthAccount", "User", "1:1"),
        ("User", "SellerProfile", "1:0..1"),
        ("User", "Order", "1:M"),
        ("Order", "OrderItem", "1:M"),
        ("Order", "Payment", "1:M"),
    ]
    for a, b, label in links:
        ax, ay = centers[a]
        bx, by = centers[b]
        doc.line(ax, ay, bx, by, 0.45, (0.25, 0.25, 0.25))
        doc.text((ax + bx) / 2, (ay + by) / 2 + 3, label, 7.2, FONT_ITALIC)


def draw_module_diagram(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    groups = [
        ("Frontend", ["User App", "Seller Dashboard", "Superadmin", "Session Analytics"], x + 8),
        ("Backend", ["API Gateway", "Auth", "Product", "Order", "Payment", "Session"], x + 150),
        ("Data", ["MySQL", "MongoDB", "Redis", "Typesense", "Kafka/RabbitMQ"], x + 292),
    ]
    for title, items, gx in groups:
        doc.rect(gx, y - 330, 124, 310, stroke=(0.20, 0.20, 0.20), fill=(0.985, 0.985, 0.985), width=0.55)
        doc.text_center(gx, y - 38, 124, title, 10.5, FONT_BOLD)
        yy = y - 70
        centers = []
        for item in items:
            centers.append(box(doc, gx + 12, yy, 100, 32, item, (0.93, 0.96, 1.0), 8.1))
            yy -= 43
        if title != "Data":
            for c in centers:
                doc.arrow(c[0] + 50, c[1], c[0] + 78, c[1])
    doc.arrow(x + 132, y - 175, x + 150, y - 175)
    doc.arrow(x + 274, y - 175, x + 292, y - 175)


def draw_auth_flow(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    lanes = [("Client", x + 20), ("Auth HTTP", x + 132), ("Usecase", x + 244), ("MySQL/Redis", x + 356)]
    for name, lx in lanes:
        box(doc, lx, y - 20, 82, 28, name, (0.96, 0.94, 0.88), 8.5)
        doc.line(lx + 41, y - 50, lx + 41, y - 340, 0.3, (0.55, 0.55, 0.55))
    steps = [
        (0, 1, "POST /login", 78),
        (1, 2, "Validate DTO", 118),
        (2, 3, "Check password hash", 158),
        (3, 2, "Account OK", 198),
        (2, 3, "Store refresh token", 238),
        (2, 1, "JWT + session_id", 278),
        (1, 0, "200 OK", 318),
    ]
    for a, b, label, dy in steps:
        x1 = lanes[a][1] + 41
        x2 = lanes[b][1] + 41
        yy = y - dy
        doc.arrow(x1, yy, x2, yy)
        doc.text(min(x1, x2) + 8, yy + 5, label, 7.8, FONT_REG)


def draw_session_flow(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    sdk = box(doc, x + 15, y - 30, 88, 40, "Frontend SDK", (0.90, 0.96, 0.92))
    gw = box(doc, x + 155, y - 30, 88, 40, "API Gateway", (0.90, 0.93, 0.98))
    session = box(doc, x + 295, y - 30, 100, 40, "Session Service", (0.90, 0.93, 0.98))
    redis = box(doc, x + 74, y - 155, 88, 40, "Redis Active Sessions", (0.98, 0.94, 0.92), 8.2)
    mongo = box(doc, x + 204, y - 155, 88, 40, "Mongo Events", (0.98, 0.94, 0.92), 8.2)
    mq = box(doc, x + 334, y - 155, 88, 40, "Queue Events", (0.98, 0.94, 0.92), 8.2)
    dash = box(doc, x + 155, y - 285, 120, 42, "Analytics Dashboard", (0.90, 0.96, 0.92), 8.5)
    doc.arrow(sdk[0] + 44, sdk[1], gw[0] - 44, gw[1])
    doc.arrow(gw[0] + 44, gw[1], session[0] - 50, session[1])
    for target in [redis, mongo, mq]:
        doc.arrow(session[0], session[1] - 20, target[0], target[1] + 20)
    doc.arrow(dash[0], dash[1] + 21, session[0] - 20, session[1] - 20)
    doc.text(x + 70, y - 86, "POST /api/v1/sessions/events", 8, FONT_ITALIC)
    doc.text(x + 155, y - 258, "GET /api/v1/analytics/live", 8, FONT_ITALIC)


def draw_deployment(doc: ReportDoc, x: float, y: float, w: float, h: float) -> None:
    internet = box(doc, x + 170, y - 16, 90, 34, "Internet Users", (0.96, 0.94, 0.88))
    cdn = box(doc, x + 50, y - 83, 90, 34, "CDN / Static", (0.90, 0.96, 0.92))
    waf = box(doc, x + 170, y - 83, 90, 34, "WAF", (0.96, 0.94, 0.88))
    lb = box(doc, x + 290, y - 83, 90, 34, "Load Balancer", (0.96, 0.94, 0.88))
    ingress = box(doc, x + 170, y - 150, 100, 34, "K8s Ingress", (0.90, 0.93, 0.98))
    gateway = box(doc, x + 170, y - 215, 100, 34, "API Gateway Pods", (0.90, 0.93, 0.98))
    core = box(doc, x + 45, y - 300, 130, 48, "Core Services\nAuth Product Order Payment", (0.93, 0.96, 1.0), 8.0)
    data = box(doc, x + 205, y - 300, 110, 48, "Managed Data\nMySQL Mongo Redis", (0.98, 0.94, 0.92), 8.0)
    obs = box(doc, x + 345, y - 300, 88, 48, "Monitoring\nLogs Metrics", (0.94, 0.94, 0.98), 8.0)
    doc.arrow(internet[0], internet[1] - 17, waf[0], waf[1] + 17)
    doc.arrow(internet[0] - 20, internet[1] - 17, cdn[0], cdn[1] + 17)
    doc.arrow(waf[0] + 45, waf[1], lb[0] - 45, lb[1])
    doc.arrow(lb[0] - 45, lb[1] - 17, ingress[0] + 50, ingress[1] + 17)
    doc.arrow(ingress[0], ingress[1] - 17, gateway[0], gateway[1] + 17)
    for target in [core, data, obs]:
        doc.arrow(gateway[0], gateway[1] - 17, target[0], target[1] + 24)


def add_title_page(doc: ReportDoc) -> None:
    doc.mark("Title Page", 0)
    doc.rect(76, 175, PAGE_W - 152, 490, stroke=(0, 0, 0), fill=None, width=1.2)
    doc.text_center(86, 625, PAGE_W - 172, "GUIDELINES BASED MCA MAJOR PROJECT REPORT", 14, FONT_BOLD)
    doc.text_center(86, 578, PAGE_W - 172, "For", 12, FONT_BOLD)
    doc.text_center(86, 538, PAGE_W - 172, "Major Project", 19, FONT_BOLD)
    doc.text_center(86, 502, PAGE_W - 172, "(23ONMCR-753)", 15, FONT_BOLD)
    doc.text_center(86, 455, PAGE_W - 172, PROJECT_TITLE, 17, FONT_BOLD)
    doc.text_center(86, 394, PAGE_W - 172, "Submitted in partial fulfillment of the requirements for the programme", 11.5, FONT_REG)
    doc.text_center(86, 365, PAGE_W - 172, "Master of Computer Applications", 18, FONT_BOLD)
    doc.text_center(86, 330, PAGE_W - 172, "Fourth Semester", 13.5, FONT_BOLD)
    doc.text(114, 284, "Submitted By: ______________________________", 11.5, FONT_REG)
    doc.text(114, 258, "Enrollment / Roll No.: _____________________", 11.5, FONT_REG)
    doc.text(114, 232, "Project Guide: _____________________________", 11.5, FONT_REG)
    doc.text_center(86, 196, PAGE_W - 172, "CENTRE FOR DISTANCE & ONLINE EDUCATION", 13, FONT_BOLD)
    doc.text_center(86, 173, PAGE_W - 172, "CHANDIGARH UNIVERSITY", 13, FONT_BOLD)


def add_certificate(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Certificate", 0)
    doc.heading("Certificate", 1)
    doc.paragraph(
        "This is to certify that the major project report entitled "
        f"'{PROJECT_TITLE}' has been prepared and submitted by "
        "______________________________ in partial fulfillment of the requirements "
        "for the Master of Computer Applications programme of Chandigarh University."
    )
    doc.paragraph(
        "The work presented in this report has been carried out as an individual "
        "major project. The report follows the official Major Project guidelines "
        "provided for course code 23ONMCR-753. The project work is based on the "
        "submitted repository documents, source code, database design files, and "
        "implementation notes."
    )
    doc.paragraph(
        "To the best of my knowledge, the project report gives a clear and honest "
        "description of the system design and the current implemented modules."
    )
    doc.y -= 40
    doc.text(MARGIN_L, doc.y, "Project Guide Signature: __________________________", 12, FONT_REG)
    doc.y -= 36
    doc.text(MARGIN_L, doc.y, "Name of Guide: ___________________________________", 12, FONT_REG)
    doc.y -= 36
    doc.text(MARGIN_L, doc.y, "Date: ____________________", 12, FONT_REG)
    doc.y -= 36
    doc.text(MARGIN_L, doc.y, "Place: ___________________", 12, FONT_REG)


def add_declaration(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Declaration", 0)
    doc.heading("Declaration", 1)
    doc.paragraph(
        "I hereby declare that the major project report entitled "
        f"'{PROJECT_TITLE}' is based on my project documents and source code. "
        "The project has been prepared as an individual academic submission for "
        "the MCA Major Project course."
    )
    doc.paragraph(
        "The information written in this report has been taken from the uploaded "
        "project guideline PDF and the project files present in the repository. "
        "Where personal details such as student name, enrollment number, and guide "
        "name were not available, blank fields have been kept so that they can be "
        "filled with correct information."
    )
    doc.paragraph(
        "I also declare that no unsupported claim has been added as a completed "
        "implementation. Modules that are documented as design or future work are "
        "described in the same way in this report."
    )
    doc.y -= 60
    doc.text(MARGIN_L, doc.y, "Student Signature: __________________________", 12, FONT_REG)
    doc.y -= 36
    doc.text(MARGIN_L, doc.y, "Student Name: ______________________________", 12, FONT_REG)
    doc.y -= 36
    doc.text(MARGIN_L, doc.y, "Date: ____________________", 12, FONT_REG)


def add_acknowledgement(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Acknowledgement", 0)
    doc.heading("Acknowledgement", 1)
    doc.paragraph(
        "I would like to express my sincere thanks to Chandigarh University and "
        "the Centre for Distance & Online Education for giving me the opportunity "
        "to complete this MCA Major Project."
    )
    doc.paragraph(
        "I am thankful to my project guide, faculty members, and evaluators for "
        "their support and guidance. Their direction helped me understand how a "
        "major project should be planned, documented, implemented, and tested."
    )
    doc.paragraph(
        "I also thank everyone who helped me directly or indirectly during the "
        "preparation of this project. This work helped me learn practical software "
        "engineering concepts such as microservices, authentication, database "
        "design, API design, testing, monitoring, and scalable deployment."
    )
    doc.paragraph(
        "Finally, I thank my family and friends for their encouragement during "
        "the project work."
    )


def add_abstract(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Abstract", 0)
    doc.heading("Abstract", 1)
    doc.paragraph(
        f"The project titled '{PROJECT_TITLE}' is a scalable e-commerce platform "
        "designed for high traffic online shopping use cases. The system is planned "
        "as a microservices-based platform where each major business area, such as "
        "authentication, user profile, product catalog, cart, order, payment, search, "
        "CMS, session analytics, notification, and superadmin, is handled by a "
        "separate service."
    )
    doc.paragraph(
        "The project uses Go for backend service development and React with "
        "TypeScript for frontend dashboard development. MySQL is used for strongly "
        "structured and transactional data such as authentication, orders, payments, "
        "CMS, and admin records. MongoDB is used for flexible and high-volume data "
        "such as product documents, cart documents, wishlist records, recommendation "
        "data, session events, and notification payloads. Redis is used for fast "
        "temporary data such as active sessions, OTP rate limits, carts, and cache. "
        "Typesense is planned for search, while Kafka or RabbitMQ is planned for "
        "event-driven communication."
    )
    doc.paragraph(
        "The current repository contains full architecture documentation, database "
        "schema design, API references, task implementation guides, a Go-based Auth "
        "Service implementation, and a React TypeScript Session Analytics Dashboard. "
        "Other services are described as part of the complete platform design and "
        "future implementation plan. The report explains the objectives, requirements, "
        "SDLC, system design, implementation details, testing approach, applications, "
        "conclusion, and bibliography in simple academic language."
    )
    doc.paragraph(
        "The main goal of this project is to show how a modern e-commerce system can "
        "be designed for modular development, security, observability, and future "
        "scalability."
    )


def add_intro(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Introduction", 0)
    doc.heading("Introduction", 1)
    doc.heading("1.1 Project Overview", 2)
    doc.paragraph(
        "E-commerce platforms are used by buyers, sellers, administrators, and "
        "support teams at the same time. A small online store can be built as a "
        "single application, but a high-performance platform needs a design that can "
        "grow safely. This project presents a scalable e-commerce platform where "
        "different parts of the system are divided into separate services."
    )
    doc.paragraph(
        "The project topic was inferred from the repository README and documents as "
        f"'{PROJECT_TITLE}'. The system is not only a shopping website idea. It also "
        "includes seller tools, superadmin tools, session analytics, payment flow, "
        "database design, DevOps planning, logging, monitoring, and security."
    )
    doc.heading("1.2 Source Documents Studied", 2)
    for item in [
        "Official MCA Major Project guideline PDF named Project Work-Assignment 1.pdf.",
        "README.md describing the project topic.",
        "Architecture, microservice, database, security, payment, session, frontend, DevOps, monitoring, and developer guide documents in the docs folder.",
        "Relational and MongoDB database design documents in the database folder.",
        "Master API reference in api/master-api.json.",
        "Auth Service source code, migrations, use cases, repositories, and tests.",
        "Session Analytics Dashboard React TypeScript source code and tests.",
        "Task-wise implementation notes under TaskImplementation.",
    ]:
        doc.bullet(item)
    doc.heading("1.3 Objectives", 2)
    objectives = [
        "To design a high-performance e-commerce platform using a microservices architecture.",
        "To separate major business features into independent modules with clear ownership.",
        "To design secure authentication, authorization, OTP, token, and role management.",
        "To design catalog, cart, order, payment, seller CMS, superadmin, search, and analytics workflows.",
        "To use proper database choices for different data types instead of using one database for all data.",
        "To support future scalability through Redis caching, message queues, Kubernetes, and observability tools.",
        "To implement selected working modules in Go and React TypeScript.",
        "To prepare proper academic documentation following the MCA Major Project guidelines.",
    ]
    for item in objectives:
        doc.bullet(item)
    doc.heading("1.4 Problem Statement", 2)
    doc.paragraph(
        "A normal e-commerce application can become difficult to manage when the "
        "number of users, products, orders, and sellers increases. If all features "
        "are kept inside one large codebase and one shared database, changes in one "
        "area can affect other areas. The system can become slow, risky to deploy, "
        "and difficult to scale."
    )
    doc.paragraph(
        "The problem addressed by this project is to design an e-commerce platform "
        "that can handle different business areas independently. The platform should "
        "support secure login, product browsing, cart, checkout, payments, seller "
        "management, admin control, session analytics, monitoring, and future growth."
    )
    doc.heading("1.5 Scope of the Project", 2)
    doc.paragraph(
        "The scope of the project includes both system design and selected practical "
        "implementation. The documents define the full platform architecture. The "
        "current code implementation mainly includes the Auth Service and the "
        "Session Analytics Dashboard. This report describes planned modules and "
        "implemented modules separately so that the report remains honest and clear."
    )
    doc.table(
        ["Area", "Included Scope"],
        [
            ["Architecture", "API Gateway, microservices, databases, Redis, queue, Kubernetes, monitoring."],
            ["Backend implementation", "Go Auth Service with password, OTP, token, role, session-link, repositories, and migrations."],
            ["Frontend implementation", "React TypeScript Session Analytics Dashboard shell with filters and metric cards."],
            ["Database design", "MySQL tables, MongoDB collections, Typesense schema, indexes, and ownership rules."],
            ["Security", "JWT, refresh token rotation, OTP hashing, RBAC, rate limiting, audit logging, privacy controls."],
            ["Deployment planning", "Docker, Kubernetes, CI/CD, external services, logging, monitoring, and scaling plan."],
        ],
        [105, 310],
    )
    doc.heading("1.6 Software Requirement Specification Summary", 2)
    doc.paragraph(
        "The Software Requirement Specification describes what the proposed system "
        "should provide. In this project, requirements are divided into functional "
        "requirements and non-functional requirements. Functional requirements explain "
        "the features. Non-functional requirements explain the quality of the system."
    )
    doc.heading("1.6.1 Functional Requirements", 3)
    doc.table(
        ["Requirement", "Simple Description"],
        [
            ["User registration and login", "The system should allow users to create account, login, logout, and manage sessions."],
            ["OTP verification", "The system should support email or phone OTP for verification and password reset."],
            ["Role-based access", "The system should allow access based on roles such as buyer, seller, admin, and superadmin."],
            ["Product catalog", "The system should support product, variant, category, brand, price, image, and inventory data."],
            ["Search and filters", "The system should allow users to search products and filter by brand, category, price, stock, and rating."],
            ["Cart and wishlist", "The system should allow add, update, remove, cart merge, coupon preview, and wishlist actions."],
            ["Checkout and order", "The system should create order from cart using idempotency and fresh stock validation."],
            ["Payment and refund", "The system should create payment intent, verify webhook, update status, and support refund flow."],
            ["Seller dashboard", "The system should allow sellers to manage products, orders, coupons, staff, and analytics."],
            ["Superadmin panel", "The system should allow admins to manage users, sellers, payments, settings, search, and audit logs."],
            ["Session analytics", "The system should track sessions, events, journeys, funnels, heatmaps, and live metrics."],
            ["Notification", "The system should send OTP, order status, payment, refund, and price-drop notifications."],
        ],
        [145, 270],
        font_size=8.1,
    )
    doc.heading("1.6.2 Non-Functional Requirements", 3)
    doc.table(
        ["Requirement", "Expected Quality"],
        [
            ["Security", "Passwords, OTPs, tokens, card data, and secrets must be protected and not logged."],
            ["Scalability", "Services should scale independently based on traffic and queue load."],
            ["Performance", "Search, cart, auth, and checkout should respond quickly under normal load."],
            ["Reliability", "Important operations should use idempotency, retries, and durable events."],
            ["Maintainability", "Code should be split into domain, usecase, repository, transport, and config layers."],
            ["Observability", "Services should produce structured logs, metrics, traces, health checks, and alerts."],
            ["Privacy", "Session analytics should mask PII, hash sensitive values, and support deletion/retention rules."],
            ["Compatibility", "Frontend should support modern browsers and responsive dashboard layouts."],
            ["Extensibility", "New services and modules should be added without changing unrelated services."],
            ["Deployment readiness", "Docker, Kubernetes, config validation, and CI/CD should be supported."],
        ],
        [145, 270],
        font_size=8.2,
    )
    doc.heading("1.7 Assumptions", 2)
    for item in [
        "The project name was taken from README.md because the user prompt contained a placeholder project name.",
        "Student name, enrollment number, guide name, and institute signature details were not provided, so blank fields are kept.",
        "Official logo image files were not provided. Therefore, the report uses a clean text header instead of logo images.",
        "The full e-commerce platform is described as the target design. The current repository implementation is limited mainly to Auth Service and Session Analytics Dashboard.",
        "The report uses simple English and avoids complex wording as requested.",
    ]:
        doc.bullet(item)
    doc.heading("1.8 Hardware Requirements", 2)
    doc.paragraph(
        "The official guideline gives a minimum hardware direction such as Pentium-IV "
        "processor or above, 2 GB RAM, 40 GB hard disk, graphics card, and other "
        "required interfaces. For this modern project, the practical development and "
        "testing setup can use the following hardware."
    )
    doc.table(
        ["Component", "Recommended Requirement", "Reason"],
        [
            ["Processor", "Intel i5/Ryzen 5 or above", "Runs backend services, frontend dev server, and local tools smoothly."],
            ["RAM", "8 GB minimum, 16 GB recommended", "Useful for Docker, database containers, tests, and browser tools."],
            ["Storage", "40 GB minimum, SSD recommended", "Stores source code, databases, packages, logs, and generated builds."],
            ["Network", "Internet connection", "Required for package setup, external APIs, and deployment work."],
            ["Display", "Standard monitor", "Required for dashboard and report verification."],
        ],
        [92, 146, 177],
        font_size=8.8,
    )
    doc.heading("1.9 Software Requirements", 2)
    doc.table(
        ["Category", "Software / Tool", "Purpose"],
        [
            ["Operating System", "Linux / Windows / macOS", "Development and testing environment."],
            ["Backend", "Go 1.24+", "Backend microservice implementation."],
            ["Frontend", "React, TypeScript, Vite, Tailwind CSS", "Session Analytics Dashboard and future web apps."],
            ["Database", "MySQL 8+, MongoDB", "Structured transactional data and flexible document/event data."],
            ["Cache / Session", "Redis", "Rate limits, active sessions, OTP counters, cart cache."],
            ["Search", "Typesense", "Fast product search and facets."],
            ["Messaging", "Kafka or RabbitMQ", "Asynchronous events and background processing."],
            ["Deployment", "Docker, Kubernetes", "Containerization and scalable deployment."],
            ["Monitoring", "Prometheus, Grafana, Loki, OpenTelemetry", "Metrics, logs, and traces."],
            ["Testing", "Go test, Vitest, Testing Library, MSW", "Backend and frontend test coverage."],
        ],
        [88, 150, 177],
        font_size=8.8,
    )
    doc.heading("1.10 Broad Areas of Application", 2)
    for item in [
        "Database Management Systems",
        "Computer Communication and Networking",
        "Mobile and Web Application Development",
        "Software Engineering",
        "Cloud Deployment and DevOps",
        "Security and Authentication",
        "Analytics and Monitoring",
    ]:
        doc.bullet(item)


def add_sdlc(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("SDLC of the Project", 0)
    doc.heading("SDLC of the Project", 1)
    doc.paragraph(
        "The project follows an iterative SDLC approach. In this method, the system "
        "is not built in one large step. First, the requirements and architecture are "
        "defined. Then modules are designed and implemented step by step. Testing and "
        "review are done after each important module."
    )
    doc.figure_page(
        "SDLC Flow Used in the Project",
        lambda d, x, y, w, h: draw_workflow(d, x, y, w, h),
        "The diagram shows a simplified implementation journey from user browsing to notification. The same step-by-step approach was used for project planning and module development.",
    )
    doc.add_page()
    doc.heading("2.1 Identification Phase", 2)
    doc.paragraph(
        "In the identification phase, the main project idea was selected: a scalable "
        "backend development project for a high-performance e-commerce platform. The "
        "tools and technologies were identified from the project documents and code. "
        "Go was selected for backend services because it is fast and suitable for "
        "network services. React TypeScript was selected for frontend dashboards. "
        "MySQL, MongoDB, Redis, Typesense, and Kafka/RabbitMQ were selected according "
        "to the type of data and workload."
    )
    doc.heading("2.2 Requirement Analysis Phase", 2)
    doc.paragraph(
        "In this phase, functional and non-functional requirements were identified. "
        "Functional requirements explain what the system should do. Non-functional "
        "requirements explain how well the system should work, such as security, "
        "performance, reliability, and scalability."
    )
    doc.table(
        ["Requirement Type", "Requirement"],
        [
            ["Functional", "User should register, login, refresh token, logout, and verify OTP."],
            ["Functional", "Buyer should browse products, search, manage cart, and complete checkout."],
            ["Functional", "Seller should manage products, inventory, orders, coupons, and analytics."],
            ["Functional", "Admin should manage users, sellers, payments, search, sessions, and audit logs."],
            ["Functional", "Session Analytics Dashboard should show live metrics using filters."],
            ["Non-functional", "System should be secure with JWT, refresh rotation, OTP hashing, RBAC, and audit logs."],
            ["Non-functional", "System should scale using microservices, caching, message queue, and Kubernetes."],
            ["Non-functional", "System should provide logs, metrics, traces, health checks, and alerts."],
        ],
        [120, 295],
    )
    doc.heading("2.3 Analysis Phase", 2)
    doc.paragraph(
        "The system was analyzed as a set of independent services. Each service owns "
        "its own data. For example, Auth Service owns credentials and tokens, Product "
        "Service owns product catalog data, Order Service owns orders, and Payment "
        "Service owns payment records. Other services should not directly read or "
        "write a service's database. They should communicate through APIs or events."
    )
    doc.heading("2.4 Design Phase", 2)
    doc.paragraph(
        "In the design phase, the system architecture, data flow, database design, "
        "module boundaries, security controls, payment flow, session analytics, and "
        "deployment plan were prepared. The design uses API Gateway for public access, "
        "gRPC for internal service communication, and message queues for asynchronous "
        "events."
    )
    doc.heading("2.5 Implementation Phase", 2)
    doc.paragraph(
        "The implementation available in the repository focuses on two practical "
        "areas. The first is the Go Auth Service, which includes HTTP handlers, use "
        "cases, password hashing, token issue and verification, OTP flow, RBAC, MySQL "
        "repositories, Redis OTP rate repository, outbox events, and migrations. The "
        "second is the React TypeScript Session Analytics Dashboard, which includes "
        "layout, filters, metric cards, API mapping, validation, error states, and "
        "unit/component tests."
    )
    doc.heading("2.6 Testing Phase", 2)
    doc.paragraph(
        "Testing is planned at different levels. Backend testing includes unit tests "
        "for use cases, token security tests, role security tests, password tests, OTP "
        "tests, and HTTP security tests. Frontend testing includes helper tests, API "
        "mapping tests, component tests, and dashboard page tests. Integration and "
        "end-to-end tests are planned for full workflows such as signup, login, search, "
        "cart, checkout, payment success, seller product creation, and admin refund "
        "review."
    )
    doc.heading("2.7 Deployment and Maintenance Phase", 2)
    doc.paragraph(
        "Deployment is planned using Docker and Kubernetes. Services run as containers. "
        "Kubernetes manages service discovery, scaling, health checks, and rolling "
        "deployments. Monitoring uses Prometheus, Grafana, Loki, and OpenTelemetry. "
        "Maintenance includes schema migrations, security updates, monitoring alerts, "
        "backup planning, and future feature development."
    )
    doc.heading("2.8 SDLC Mapping Table", 2)
    doc.table(
        ["SDLC Phase", "Project Work"],
        [
            ["Identification", "Selected scalable e-commerce platform and identified tools."],
            ["Requirement Analysis", "Prepared functional and non-functional requirements."],
            ["System Design", "Prepared microservice architecture, database design, API design, and security design."],
            ["Implementation", "Implemented Auth Service and Session Analytics Dashboard modules."],
            ["Testing", "Added backend and frontend tests for implemented modules."],
            ["Deployment Planning", "Documented Docker, Kubernetes, CI/CD, monitoring, and scaling approach."],
            ["Maintenance", "Prepared future enhancement and production readiness plan."],
        ],
        [120, 295],
    )


def add_design(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Design", 0)
    doc.heading("Design", 1)
    doc.heading("3.1 Overall System Architecture", 2)
    doc.paragraph(
        "The platform is designed as a microservices system. Public requests from "
        "web apps go through the Load Balancer, Kubernetes Ingress, and API Gateway. "
        "The API Gateway validates requests, checks authentication where required, "
        "adds request IDs, performs rate limiting, and routes requests to the correct "
        "backend service."
    )
    doc.figure_page(
        "System Architecture",
        draw_system_architecture,
        "The diagram shows the major frontend apps, gateway, backend services, databases, cache, search engine, and message queue. The API Gateway is the main entry point for public traffic.",
    )
    doc.add_page()
    doc.heading("3.2 Main Modules", 2)
    doc.table(
        ["Module", "Main Responsibility"],
        [
            ["API Gateway", "Public REST entry point, routing, validation, auth checks, rate limiting, and error mapping."],
            ["Auth Service", "Login, password validation, JWT, refresh tokens, OTP, RBAC, and logout."],
            ["User Service", "User profile, addresses, seller profile, KYC metadata, and user status."],
            ["Product Service", "Catalog, variants, attributes, inventory, categories, and product publishing."],
            ["Cart Service", "User and guest carts, item quantity, coupon preview, and cart merge."],
            ["Wishlist Service", "Wishlist items, move to cart, price drop and availability tracking."],
            ["Order Service", "Cart to order conversion, order lifecycle, fulfillment, and cancellation."],
            ["Payment Service", "Payment intents, provider webhooks, refunds, reconciliation, and audit."],
            ["Search Service", "Typesense product search, autocomplete, facets, synonyms, and reindexing."],
            ["CMS Service", "Seller product workflow, coupons, campaigns, seller settings, and analytics."],
            ["Session Service", "Session events, user journey, live metrics, funnels, heatmaps, and privacy."],
            ["Notification Service", "Email, SMS, push notifications, templates, delivery logs, and retries."],
            ["Superadmin Service", "User, seller, order, payment, search, settings, sessions, and audit controls."],
        ],
        [110, 305],
        font_size=8.6,
    )
    doc.figure_page(
        "Module Diagram",
        draw_module_diagram,
        "The module diagram groups the system into frontend apps, backend services, and data/infra components. This makes the ownership of each part easy to understand.",
    )
    doc.add_page()
    doc.heading("3.3 Workflow Diagram", 2)
    doc.paragraph(
        "A typical e-commerce workflow starts when a user searches or browses a "
        "product. The user adds an item to the cart and starts checkout. The system "
        "verifies the user session, creates an order, reserves inventory, creates a "
        "payment intent, waits for the provider webhook, marks the order as paid, "
        "commits inventory, and sends notification."
    )
    doc.figure_page(
        "Checkout and Payment Workflow",
        draw_workflow,
        "The workflow explains how product browsing, cart, checkout, auth, inventory, payment, webhook, order update, and notification are connected.",
    )
    doc.add_page()
    doc.heading("3.4 Use Case Diagram", 2)
    doc.paragraph(
        "The use case diagram shows the main actors and their interactions with the "
        "platform. Buyer uses shopping features, seller uses product and analytics "
        "features, admin uses platform management features, and payment provider "
        "interacts with checkout and refund flows."
    )
    doc.figure_page(
        "Use Case Diagram",
        draw_use_case,
        "The diagram shows buyer, seller, admin, and payment provider as actors. It also shows the main user-facing and admin-facing use cases.",
    )
    doc.add_page()
    doc.heading("3.5 Data Flow Diagram", 2)
    doc.paragraph(
        "Data flow design explains how data moves between users, the platform, "
        "external providers, and data stores. In this system, users and dashboards "
        "do not directly access databases. They use the API Gateway and services."
    )
    doc.figure_page(
        "Data Flow Diagram - Level 0",
        draw_dfd,
        "The DFD shows external users, the e-commerce platform, payment provider, notification provider, and main data stores. It keeps the view simple for beginner understanding.",
    )
    doc.add_page()
    doc.heading("3.6 Database Design", 2)
    doc.paragraph(
        "The platform uses database-per-service ownership. This means every service "
        "owns its own database or collections. No service is allowed to directly use "
        "another service's database tables. This rule keeps the system modular and "
        "reduces hidden coupling."
    )
    doc.table(
        ["Service", "Database", "Reason"],
        [
            ["Auth", "MySQL + Redis", "Credentials and tokens need consistency; Redis supports rate limits and short-lived state."],
            ["User", "MySQL", "Profiles, addresses, and seller KYC are structured relational data."],
            ["Product", "MongoDB", "Product attributes and variants are flexible by category."],
            ["Cart", "MongoDB + Redis", "Cart document is flexible; Redis helps hot cart reads."],
            ["Wishlist", "MongoDB", "Wishlist is document-friendly and read-heavy."],
            ["Order", "MySQL", "Orders need transaction and audit records."],
            ["Payment", "MySQL", "Payment and refund data need strong audit and consistency."],
            ["Search", "Typesense", "Fast text search, filters, facets, and sorting."],
            ["Session", "MongoDB + Redis", "High-volume flexible events and active sessions."],
            ["Notification", "MongoDB", "Templates and provider payloads vary."],
            ["Superadmin", "MySQL", "Permissions, settings, tasks, and audit logs are structured."],
        ],
        [80, 105, 230],
        font_size=8.5,
    )
    doc.figure_page(
        "Entity Relationship Diagram",
        draw_er,
        "The ER diagram focuses on important relational entities from authentication, user, order, and payment areas. It shows one-to-one and one-to-many relations.",
    )
    doc.add_page()
    doc.heading("3.7 Important MySQL Entities", 2)
    doc.paragraph(
        "The relational schema includes separate databases such as auth_db, user_db, "
        "order_db, payment_db, cms_db, and admin_db. The implemented Auth Service "
        "migrations define auth_accounts, credentials, refresh_tokens, otp_challenges, "
        "role_assignments, and auth_outbox_events."
    )
    doc.table(
        ["Entity", "Purpose"],
        [
            ["auth_accounts", "Stores account id, email, phone, verification flags, and account status."],
            ["credentials", "Stores password hash, algorithm, failed attempts, and lockout time."],
            ["refresh_tokens", "Stores hashed refresh tokens with session id, expiry, and revoke time."],
            ["otp_challenges", "Stores OTP challenge id, target, channel, purpose, hash, attempts, and expiry."],
            ["role_assignments", "Stores account role, scope, assigned by, assigned at, and revoke status."],
            ["auth_outbox_events", "Stores auth events for async publishing to session/fraud systems."],
            ["orders", "Stores order id, user id, status, amount fields, address snapshot, and payment id."],
            ["order_items", "Stores item snapshot, seller id, product id, variant id, quantity, and fulfillment status."],
            ["payments", "Stores payment status, provider references, amount, and audit details."],
        ],
        [120, 295],
        font_size=8.6,
    )
    doc.heading("3.8 MongoDB Collections", 2)
    doc.paragraph(
        "MongoDB is used where records are flexible and document-based. Product "
        "documents can contain category-specific attributes and variant arrays. "
        "Session events are stored as separate documents because events can be very "
        "high in number."
    )
    doc.table(
        ["Collection", "Service", "Purpose"],
        [
            ["products", "Product", "Product details, attributes, images, variants, price, stock, rating."],
            ["categories", "Product", "Category tree and display order."],
            ["carts", "Cart", "Active cart items, coupon, totals, expiry."],
            ["wishlists", "Wishlist", "Saved products and last known availability."],
            ["user_interactions", "Recommendation", "Product view and behavior events for recommendation."],
            ["sessions", "Session", "Session id, anonymous id, user id, device, geo, start, last seen."],
            ["session_events", "Session", "Page view, click, search, add to cart, checkout events."],
            ["heatmap_points", "Session", "Aggregated click or scroll positions."],
            ["notification_templates", "Notification", "Email/SMS/push template definitions."],
            ["notification_deliveries", "Notification", "Delivery status, provider id, attempts, and payload."],
        ],
        [115, 90, 210],
        font_size=8.5,
    )
    doc.heading("3.9 API Design", 2)
    doc.paragraph(
        "The platform exposes REST APIs through the API Gateway. Internal services "
        "can use gRPC when typed and fast service-to-service calls are needed. Public "
        "browser APIs use paths like /api/v1/auth/login and /api/v1/analytics/live."
    )
    doc.table(
        ["API", "Method", "Purpose"],
        [
            ["/api/v1/auth/login", "POST", "Login user and return access token, refresh token, and session id."],
            ["/api/v1/auth/refresh", "POST", "Rotate refresh token and issue a new access token."],
            ["/api/v1/auth/logout", "POST", "Revoke refresh token and end session."],
            ["/api/v1/auth/otp/send", "POST", "Create OTP challenge and send OTP through notification provider."],
            ["/api/v1/auth/otp/verify", "POST", "Verify OTP for signup, login, or password reset."],
            ["/api/v1/search", "GET", "Search products with query, filters, facets, and sorting."],
            ["/api/v1/orders/checkout", "POST", "Create order from cart and initiate payment."],
            ["/api/v1/webhooks/payments/{provider}", "POST", "Receive and verify payment provider webhook."],
            ["/api/v1/analytics/live", "GET", "Fetch live session metrics for dashboard filters."],
            ["/api/v1/analytics/heatmaps", "GET", "Fetch heatmap data for selected route and segment."],
        ],
        [170, 55, 190],
        font_size=8.1,
    )
    doc.heading("3.10 Security Design", 2)
    doc.paragraph(
        "Security is treated as a foundation of the platform. The Auth Service handles "
        "password hashing, JWT token issue, refresh token rotation, OTP challenges, "
        "role assignment, and logout. Sensitive values such as OTP, refresh tokens, "
        "passwords, card data, and raw secrets must not be logged."
    )
    doc.figure_page(
        "Authentication Flow",
        draw_auth_flow,
        "The flow shows how login passes through HTTP transport, usecase layer, MySQL/Redis repositories, and returns token/session details to the client.",
    )
    doc.add_page()
    doc.heading("3.11 Session Analytics Design", 2)
    doc.paragraph(
        "The Session Management Service is independent from basic login. It tracks "
        "anonymous and logged-in sessions, page views, clicks, searches, product "
        "views, add-to-cart events, checkout steps, payment result events, journeys, "
        "funnels, heatmaps, and live metrics. Privacy rules say that raw IP, OTP, "
        "passwords, card data, and private text fields must not be collected."
    )
    doc.figure_page(
        "Session Analytics Data Flow",
        draw_session_flow,
        "The diagram shows frontend event collection, gateway forwarding, session service processing, Redis active-session updates, MongoDB event storage, queue publishing, and dashboard metric reading.",
    )
    doc.add_page()
    doc.heading("3.12 Privacy and Data Retention Design", 2)
    doc.paragraph(
        "Session analytics is useful, but it can also become sensitive if personal "
        "or behavioral data is shown without control. Therefore, the design includes "
        "privacy rules. Raw IP should not be stored unless a strict security reason "
        "exists. IP hash, user id, anonymous id, and session id should be masked by "
        "default in admin screens. Passwords, OTP, card data, private text fields, "
        "and keystrokes should never be collected in analytics events."
    )
    doc.table(
        ["Data", "Privacy Rule", "Retention Direction"],
        [
            ["Raw session events", "Avoid PII and sensitive fields.", "TTL based retention, for example 30 to 90 days."],
            ["Journey summaries", "Use masked user/session ids.", "Medium retention, for example 180 to 365 days."],
            ["Heatmap points", "Store aggregated coordinates, not private text.", "Longer retention because it is aggregate data."],
            ["Analytics aggregates", "Keep identity-free counts and rates.", "Long-term storage allowed if no direct identity exists."],
            ["IP address", "Do not show raw IP. Use hash if needed.", "Keep only as long as security policy requires."],
            ["Deletion requests", "Delete or anonymize linkable session data.", "Audit request and result safely."],
        ],
        [115, 160, 140],
        font_size=8.2,
    )
    doc.heading("3.13 Caching and Scalability Design", 2)
    doc.paragraph(
        "The platform separates hot data from durable data. Redis is used for values "
        "that need fast access, such as active sessions, cart count, OTP retry counters, "
        "rate limit buckets, product summary cache, seller settings cache, and platform "
        "settings cache. Durable data remains in MySQL or MongoDB."
    )
    doc.table(
        ["Data", "Cache", "Typical TTL"],
        [
            ["Product summary", "Redis or CDN edge where safe", "5 to 15 minutes"],
            ["Category tree", "Redis", "30 minutes"],
            ["Cart count", "Redis", "5 minutes"],
            ["Search autocomplete", "Redis", "1 to 5 minutes"],
            ["Seller settings", "Redis", "10 minutes"],
            ["Platform settings", "Redis", "5 minutes"],
            ["Recommendations", "Redis", "15 minutes"],
            ["Active sessions", "Redis", "Session inactivity window"],
        ],
        [130, 170, 115],
        font_size=8.3,
    )
    doc.heading("3.14 Observability Design", 2)
    doc.paragraph(
        "Observability helps developers and operations teams understand what is "
        "happening inside the platform. The project documents recommend structured "
        "JSON logs, Prometheus metrics, distributed tracing through OpenTelemetry, "
        "and dashboards or alerts through Grafana."
    )
    doc.table(
        ["Observability Area", "Examples"],
        [
            ["Logs", "timestamp, level, service, environment, request_id, trace_id, route, error_code, latency_ms."],
            ["Metrics", "request count, latency p95/p99, error rate, DB latency, Redis latency, queue lag."],
            ["Business metrics", "signup rate, product views, add to cart rate, checkout started, payment success rate."],
            ["Traces", "search request, checkout, payment webhook, seller product publish, admin refund review."],
            ["Alerts", "gateway 5xx, checkout latency, payment webhook failures, queue lag, DB CPU, Redis memory."],
        ],
        [140, 275],
        font_size=8.2,
    )
    doc.heading("3.15 Payment Design", 2)
    doc.paragraph(
        "Payment Service abstracts the external payment provider. The final payment "
        "status is not decided by the frontend success page. It is decided by the "
        "verified provider webhook. This prevents false success states and makes "
        "financial records more reliable."
    )
    doc.table(
        ["Payment State", "Meaning"],
        [
            ["initiated", "Payment intent was created."],
            ["requires_action", "Provider needs extra user action such as OTP or bank confirmation."],
            ["authorized", "Provider authorized the payment."],
            ["captured", "Payment is successful and money is captured."],
            ["failed", "Payment failed and order can move to payment_failed."],
            ["retry_allowed", "User can retry payment for the same order."],
            ["partially_refunded", "Some amount was refunded."],
            ["refunded", "Full amount was refunded."],
        ],
        [130, 285],
    )
    doc.heading("3.16 Deployment Design", 2)
    doc.paragraph(
        "The deployment design uses Docker images and Kubernetes manifests. Public "
        "traffic reaches the cloud load balancer, WAF, ingress controller, and API "
        "Gateway pods. Internal services are exposed through ClusterIP services. "
        "Observability tools collect logs, metrics, and traces."
    )
    doc.figure_page(
        "Deployment Architecture",
        draw_deployment,
        "The deployment diagram shows the path from internet users to Kubernetes services, managed data stores, and monitoring tools.",
    )


def add_implementation(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Coding & Implementation", 0)
    doc.heading("Coding & Implementation", 1)
    doc.heading("4.1 Repository Structure", 2)
    doc.paragraph(
        "The repository is organized into documentation, API reference, database "
        "design, backend services, frontend apps, and task implementation notes. This "
        "structure follows the project documentation and keeps each concern separate."
    )
    doc.code_block([
        "Ecommerce/",
        "  README.md",
        "  Project Work-Assignment 1.pdf",
        "  docs/",
        "  api/master-api.json",
        "  database/draw.sql",
        "  database/mongodb-schema-design.md",
        "  backend/services/auth-service/",
        "  frontend/session-analytics-dashboard/",
        "  TaskImplementation/",
        "  reports/",
    ])
    doc.heading("4.2 Backend Auth Service Implementation", 2)
    doc.paragraph(
        "The Auth Service is implemented in Go. The service starts from cmd/server/"
        "main.go. It loads configuration, connects to MySQL, connects to Redis, "
        "initializes password hashing, OTP generator, OTP hasher, notification client, "
        "repositories, JWT issuer/verifier, session linker, use cases, HTTP handler, "
        "router, and HTTP server."
    )
    doc.table(
        ["Layer", "Files / Package", "Responsibility"],
        [
            ["cmd/server", "main.go", "Application bootstrap and dependency wiring."],
            ["config", "internal/config", "Load HTTP, DB, Redis, token, OTP, password, and session-link settings."],
            ["domain", "internal/domain", "Account, credential, token, OTP, role, outbox, and domain errors."],
            ["usecase", "internal/usecase", "Login, token, password, OTP, role, reset password, and signup workflows."],
            ["repository", "internal/repository", "MySQL and Redis repository implementations."],
            ["security", "internal/security", "Password hashing, token issue/verify, OTP generator and hasher."],
            ["transport", "internal/transport/http", "HTTP routes, DTO mapping, middleware, and error responses."],
            ["events", "internal/events", "Outbox worker and event publishing."],
            ["sessionlink", "internal/sessionlink", "Session-link events and privacy hashing."],
        ],
        [78, 115, 222],
        font_size=8.1,
    )
    doc.heading("4.3 Implemented Auth Routes", 2)
    doc.paragraph(
        "The HTTP router registers public auth routes, internal service routes, role "
        "management routes, and JWKS route. The health route /healthz returns status ok."
    )
    doc.table(
        ["Route", "Purpose"],
        [
            ["/api/v1/auth/login", "Login and receive user, token pair, and session id."],
            ["/api/v1/auth/refresh", "Refresh access token using refresh token rotation."],
            ["/api/v1/auth/logout", "Logout and revoke refresh token."],
            ["/api/v1/auth/otp/send", "Create OTP challenge and send OTP."],
            ["/api/v1/auth/otp/verify", "Verify OTP challenge."],
            ["/api/v1/auth/password/forgot", "Start password reset OTP challenge."],
            ["/internal/v1/auth/credentials", "Create credential for an account."],
            ["/internal/v1/auth/password/verify", "Verify password internally."],
            ["/internal/v1/auth/tokens/issue", "Issue token pair internally."],
            ["/internal/v1/auth/tokens/verify", "Verify access token internally."],
            ["/internal/v1/auth/roles", "Read user roles."],
            ["/internal/v1/auth/roles/assign", "Assign roles to allowed admins."],
            ["/internal/v1/auth/roles/revoke", "Revoke roles from allowed admins."],
            ["/.well-known/jwks.json", "Expose JSON Web Key Set for token verification."],
        ],
        [190, 225],
        font_size=8.1,
    )
    doc.heading("4.4 Password and Token Implementation", 2)
    doc.paragraph(
        "The password module supports secure password hashing through a router. The "
        "project documents mention Argon2id or bcrypt. The implementation stores "
        "password hash and algorithm metadata in credentials table. Failed login "
        "attempts and lockout time are also stored."
    )
    doc.paragraph(
        "The token module issues JWT access tokens and stores refresh token hashes. "
        "Refresh token rotation is used. This means a refresh request should replace "
        "the old refresh token with a new one. If an old token is reused, it can be "
        "treated as a fraud or compromise signal."
    )
    doc.heading("4.5 OTP Implementation", 2)
    doc.paragraph(
        "The OTP flow creates an OTP challenge, stores only the OTP hash, sends OTP "
        "through the notification client, and verifies the submitted OTP against the "
        "stored hash. The policy includes expiry, maximum attempts, resend cooldown, "
        "and rate limits. Redis is used for OTP rate counters."
    )
    doc.heading("4.6 RBAC Implementation", 2)
    doc.paragraph(
        "Role-Based Access Control is used to control access to sensitive routes. "
        "Roles include buyer, seller, admin, finance_admin, catalog_admin, operations_admin, "
        "and superadmin. High-risk role mutation routes are protected by authentication "
        "and required role middleware."
    )
    doc.heading("4.7 Outbox and Session-Link Implementation", 2)
    doc.paragraph(
        "The Auth Service includes auth_outbox_events table and an outbox worker. "
        "This supports reliable asynchronous event publishing. For example, login, "
        "logout, token reuse, or session-link events can be stored first and then "
        "published to the Session Service or fraud systems."
    )
    doc.heading("4.8 Auth Database Migrations", 2)
    doc.paragraph(
        "The Auth Service migrations create and update the authentication database. "
        "The first migration creates auth_accounts, credentials, refresh_tokens, "
        "otp_challenges, and role_assignments. Later migrations add token claim metadata, "
        "OTP account foreign key, hardened role assignment fields, and outbox events."
    )
    doc.table(
        ["Migration", "Purpose"],
        [
            ["001_create_auth_tables", "Creates main auth tables."],
            ["002_add_token_claim_metadata", "Adds metadata needed for token claims."],
            ["003_add_otp_account_fk", "Hardens OTP relation with auth account."],
            ["004_harden_role_assignments", "Improves role assignment safety."],
            ["005_create_auth_outbox_events", "Creates durable event outbox table."],
        ],
        [170, 245],
    )
    doc.heading("4.9 Session Analytics Dashboard Implementation", 2)
    doc.paragraph(
        "The Session Analytics Dashboard is implemented as a Vite React TypeScript "
        "frontend app. It uses TanStack Query for server data, React Router for app "
        "routing, Tailwind CSS for styling, lucide-react for icons, date-fns for date "
        "helpers, Vitest and Testing Library for tests, and MSW for API mocking."
    )
    doc.table(
        ["File / Folder", "Purpose"],
        [
            ["src/main.tsx", "Bootstraps the React application."],
            ["src/app.tsx", "Defines providers and route structure."],
            ["src/layout/analytics-layout.tsx", "Creates dashboard layout."],
            ["src/layout/filters-bar.tsx", "Date range controls, segment filters, and refresh button."],
            ["src/features/shell/pages/analytics-overview-page.tsx", "Main overview page for live metrics."],
            ["src/features/shell/components/metric-card.tsx", "Single metric display component."],
            ["src/features/shell/components/metric-card-grid.tsx", "Responsive metric grid."],
            ["src/features/shell/components/segment-filter-panel.tsx", "Device, channel, source, and user type filters."],
            ["src/api/session-api.ts", "Maps /api/v1/analytics/live DTO to frontend domain model."],
            ["src/services/session-analytics-service.ts", "Validates request and calls analytics repository."],
            ["src/domain/analytics.ts", "Defines dashboard domain types."],
            ["src/domain/validation.ts", "Validates date range and filter request."],
        ],
        [190, 225],
        font_size=8.1,
    )
    doc.heading("4.10 Dashboard Metrics", 2)
    doc.paragraph(
        "The live metrics model includes active users now, active sessions, events "
        "per minute, sessions today, conversion rate, average session duration, bounce "
        "rate, product view to cart rate, cart to checkout rate, checkout to paid rate, "
        "and generated at timestamp."
    )
    doc.table(
        ["Metric", "Meaning"],
        [
            ["activeUsersNow", "Number of active users currently using the platform."],
            ["activeSessions", "Number of active session windows."],
            ["eventsPerMinute", "Recent event ingestion speed."],
            ["sessionsToday", "Number of sessions created today."],
            ["conversionRate", "Overall conversion percentage."],
            ["averageSessionDurationSeconds", "Average length of user session."],
            ["bounceRate", "Sessions with no deeper interaction."],
            ["productViewToCartRate", "How many product views become cart additions."],
            ["cartToCheckoutRate", "How many carts reach checkout."],
            ["checkoutToPaidRate", "How many checkouts become paid orders."],
        ],
        [170, 245],
        font_size=8.4,
    )
    doc.heading("4.11 API Mapping in Dashboard", 2)
    doc.paragraph(
        "The dashboard repository calls GET /api/v1/analytics/live with from, to, "
        "device_type, channel, source, and user_type query parameters. The response "
        "mapper accepts both snake_case and camelCase field names. This makes the "
        "frontend more tolerant during backend API evolution."
    )
    doc.code_block([
        "GET /api/v1/analytics/live",
        "Query:",
        "  from=YYYY-MM-DD",
        "  to=YYYY-MM-DD",
        "  device_type=all|desktop|mobile|tablet",
        "  channel=all|web|mobile_web|app",
        "  source=all|direct|search|paid|social|email",
        "  user_type=all|anonymous|logged_in",
    ])
    doc.heading("4.12 Implementation Limitations", 2)
    doc.paragraph(
        "The repository currently contains a strong design for many services, but not "
        "all services are fully implemented in source code. This report therefore "
        "separates documented design from implemented code. The implemented code areas "
        "are Auth Service and Session Analytics Dashboard. Product, cart, order, payment, "
        "CMS, superadmin, search, recommendation, and notification services are included "
        "as planned modules based on the project documentation."
    )
    doc.heading("4.13 Configuration Implementation", 2)
    doc.paragraph(
        "The Auth Service configuration is loaded during startup. The configuration "
        "covers HTTP address and timeouts, database DSN, Redis connection settings, "
        "password hashing policy, OTP policy, token issuer and audience, key paths, "
        "refresh token TTL, notification endpoint, RBAC settings, and session-link "
        "event publishing mode. Validating configuration at startup is important "
        "because a wrong secret, missing key, or invalid timeout can make a security "
        "service unsafe or unavailable."
    )
    doc.table(
        ["Config Area", "Example Values / Purpose"],
        [
            ["HTTP", "address, read timeout, write timeout, idle timeout."],
            ["Database", "MySQL DSN for auth_db."],
            ["Redis", "address, password, DB number, dial/read/write timeout."],
            ["Password", "hashing algorithm, memory/time cost, max failed attempts, lockout duration."],
            ["OTP", "length, expiry, attempts, resend cooldown, rate limit pepper."],
            ["Token", "issuer, audience, key id, private/public key path, access TTL, refresh TTL."],
            ["Notification", "OTP send endpoint and timeout."],
            ["Session link", "mode, auth events topic, outbox worker batch size, retry/backoff settings."],
        ],
        [120, 295],
        font_size=8.2,
    )
    doc.heading("4.14 Error Handling and Response Format", 2)
    doc.paragraph(
        "The HTTP transport layer converts use case errors into API errors. This "
        "keeps internal domain errors separate from public response messages. The "
        "handler also limits request body size, checks HTTP method, decodes JSON, "
        "normalizes request data, and writes JSON responses."
    )
    doc.code_block([
        "Success response idea:",
        "{",
        "  \"data\": { ... },",
        "  \"request_id\": \"req_123\",",
        "  \"error\": null",
        "}",
        "",
        "Error response idea:",
        "{",
        "  \"data\": null,",
        "  \"request_id\": \"req_123\",",
        "  \"error\": { \"code\": \"INVALID_REQUEST\", \"message\": \"Invalid request\" }",
        "}",
    ])
    doc.heading("4.15 Coding Standards Followed", 2)
    doc.paragraph(
        "The backend uses a clean layered structure. Domain files contain core data "
        "and rules. Use cases contain application workflows. Repository files contain "
        "database operations. Transport files only map HTTP requests and responses. "
        "This makes the code easier to test because use cases can depend on interfaces "
        "instead of direct database implementations."
    )
    doc.paragraph(
        "The frontend follows a feature-based structure. Shared types are placed in "
        "the domain folder. API mapping is kept inside api/session-api.ts. UI components "
        "are separated into metric cards, filter panel, layout, and page files. This "
        "keeps the dashboard easy to extend with future pages such as live sessions, "
        "journey explorer, funnels, heatmaps, cohorts, reports, and privacy controls."
    )


def add_testing(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Testing", 0)
    doc.heading("Testing", 1)
    doc.paragraph(
        "Testing is important because the platform handles sensitive actions such as "
        "login, OTP, role assignment, checkout, payment, refunds, and admin operations. "
        "The testing plan covers backend unit tests, security tests, frontend unit tests, "
        "component tests, API mock tests, integration tests, and future end-to-end tests."
    )
    doc.heading("5.1 Backend Testing", 2)
    doc.paragraph(
        "The Go Auth Service contains several test files for password use case, OTP "
        "use case, token use case, role security, auth use case, session-link events, "
        "password security, OTP security, token security, HTTP security, and RBAC security. "
        "These tests are focused on important security behavior."
    )
    doc.table(
        ["Test Area", "Purpose"],
        [
            ["Password tests", "Verify password hashing, password checks, reset, lockout, and failed attempts."],
            ["OTP tests", "Verify OTP generation, hashing, attempts, expiry, and verification behavior."],
            ["Token tests", "Verify JWT issue, validation, refresh token rotation, and security cases."],
            ["Role tests", "Verify role assignment, revoke, and permission checks."],
            ["HTTP security tests", "Verify methods, body limits, error mapping, auth middleware, and protected routes."],
            ["Session-link tests", "Verify auth event creation and privacy-safe session linking."],
        ],
        [135, 280],
    )
    doc.heading("5.2 Frontend Testing", 2)
    doc.paragraph(
        "The Session Analytics Dashboard includes tests for API mapping, date range "
        "helpers, formatting helpers, metric card component, segment filter panel, "
        "and analytics overview page. MSW is used for mocking API calls in tests."
    )
    doc.table(
        ["Frontend Test", "Purpose"],
        [
            ["session-api.test.ts", "Checks API response mapping for live metrics."],
            ["date-range.test.ts", "Checks date preset and custom date helper behavior."],
            ["format.test.ts", "Checks number, percent, and duration formatting."],
            ["metric-card.test.tsx", "Checks metric card display states."],
            ["segment-filter-panel.test.tsx", "Checks segment filter interactions."],
            ["analytics-overview-page.test.tsx", "Checks overview page loading, data, error, and empty states."],
        ],
        [165, 250],
    )
    doc.heading("5.3 Test Cases", 2)
    doc.table(
        ["Test Case", "Input / Action", "Expected Result"],
        [
            ["Login success", "Valid identifier and password", "User receives access token, refresh token, and session id."],
            ["Login failure", "Wrong password", "System returns auth error and increases failed attempt count."],
            ["Refresh token", "Valid refresh token", "New token pair is issued and old refresh token is rotated."],
            ["Refresh reuse", "Old refresh token reused", "System detects risky reuse and can revoke token family."],
            ["OTP send", "Valid target and purpose", "Challenge id is created and OTP send request is accepted."],
            ["OTP verify success", "Correct OTP before expiry", "Challenge is marked verified."],
            ["OTP verify failure", "Wrong OTP many times", "Attempts are counted and max attempts are enforced."],
            ["Role assign", "Allowed admin assigns role", "Role assignment is stored with audit details."],
            ["Analytics filter", "Valid date range and filters", "Dashboard fetches metrics from live analytics API."],
            ["Analytics invalid range", "Date range exceeds max days", "Dashboard shows validation message and avoids API call."],
            ["Payment webhook", "Valid provider event id and signature", "Payment status is updated idempotently."],
            ["Checkout duplicate", "Repeated idempotency key", "System prevents duplicate order/payment creation."],
        ],
        [102, 145, 168],
        font_size=7.8,
    )
    doc.heading("5.4 Integration Testing Plan", 2)
    doc.paragraph(
        "Integration testing should be done after more backend services are implemented. "
        "The main integration workflows are signup with OTP, login and refresh token, "
        "product search, add to cart, checkout, payment success, payment failure retry, "
        "seller product creation, superadmin seller approval, refund review, and analytics "
        "event ingestion."
    )
    doc.heading("5.5 Non-Functional Testing", 2)
    doc.table(
        ["Testing Type", "What to Check"],
        [
            ["Performance", "Search latency, checkout p95 latency, API Gateway throughput, and session event ingestion rate."],
            ["Security", "JWT validation, RBAC, OTP attempt limits, rate limits, webhook signatures, and secret redaction."],
            ["Reliability", "Retry behavior, dead-letter queue, idempotency, graceful shutdown, and health endpoints."],
            ["Compatibility", "Frontend behavior on desktop, tablet, mobile width, and modern browsers."],
            ["Scalability", "Horizontal scaling of stateless services and queue consumers."],
            ["Usability", "Simple dashboard filters, clear errors, readable metrics, and accessible controls."],
        ],
        [130, 285],
    )
    doc.heading("5.6 Verification Performed for This Report", 2)
    doc.paragraph(
        "The report was generated from the local repository documents and code. The "
        "PDF generation was verified using a local Python script that writes a valid "
        "PDF file and checks that the output file exists with expected page objects. "
        "Because Node.js is not usable in this environment, frontend npm tests could "
        "not be executed here. Go tests can be executed separately in the backend "
        "workspace."
    )
    doc.heading("5.7 Suggested Test Commands", 2)
    doc.paragraph(
        "The following commands are useful for validating the project in a normal "
        "developer setup. Some commands depend on local tools such as Go, Node.js, "
        "npm, and running database services."
    )
    doc.code_block([
        "# Backend Auth Service",
        "cd backend/services/auth-service",
        "go test ./...",
        "go vet ./...",
        "",
        "# Frontend Session Analytics Dashboard",
        "cd frontend/session-analytics-dashboard",
        "npm run typecheck",
        "npm run test",
        "npm run build",
        "",
        "# Report generation",
        "python3 reports/generate_mca_report.py",
    ])
    doc.heading("5.8 Acceptance Criteria", 2)
    doc.table(
        ["Area", "Acceptance Criteria"],
        [
            ["Report", "Follows guideline structure, A4 layout, Times font family, 1.5 spacing, page numbering, and required sections."],
            ["Auth Service", "Login, token refresh, logout, OTP, password reset, JWKS, and RBAC routes are registered."],
            ["Database", "Auth migrations create required tables and indexes without cross-service database dependency."],
            ["Security", "Passwords, OTPs, refresh tokens, and secrets are not stored or logged in plain form."],
            ["Dashboard", "Date range, segment filters, live metric cards, validation, loading, empty, and error states work."],
            ["API mapping", "Dashboard maps both snake_case and camelCase live metrics responses."],
            ["Testing", "Important backend and frontend behaviors have focused test files."],
            ["Documentation", "Design and implemented scope are clearly separated."],
        ],
        [95, 320],
        font_size=8.1,
    )


def add_application(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Application", 0)
    doc.heading("Application", 1)
    doc.paragraph(
        "The project can be used as the foundation for a modern online shopping "
        "platform. It is suitable for systems where buyers, sellers, and admins work "
        "at the same time and where future growth is expected."
    )
    doc.heading("6.1 Practical Applications", 2)
    for item in [
        "Online retail marketplace similar to Amazon or Flipkart style platforms.",
        "Seller dashboard for product, inventory, order, coupon, and revenue management.",
        "Superadmin panel for platform operations, user management, seller approval, refund review, and audit logs.",
        "Session analytics dashboard for live metrics, journey analysis, funnels, heatmaps, device reports, and privacy controls.",
        "Secure authentication service for any platform needing login, OTP, refresh tokens, and RBAC.",
        "Reference architecture for students learning microservices, database-per-service design, and DevOps.",
    ]:
        doc.bullet(item)
    doc.heading("6.2 Benefits", 2)
    doc.table(
        ["Benefit", "Explanation"],
        [
            ["Scalability", "Each service can scale based on its own workload."],
            ["Security", "Auth, RBAC, OTP, token rotation, rate limits, and audit logs reduce risk."],
            ["Maintainability", "Clear module boundaries make changes easier."],
            ["Performance", "Redis cache, Typesense search, and async queues support faster user experience."],
            ["Observability", "Logs, metrics, traces, and alerts help detect problems early."],
            ["Flexibility", "MongoDB supports flexible product and event documents."],
            ["Reliability", "Idempotency and outbox/event patterns reduce duplicate or lost operations."],
        ],
        [120, 295],
    )
    doc.heading("6.3 Users of the System", 2)
    doc.table(
        ["User Type", "Main Use"],
        [
            ["Buyer", "Search products, manage cart, checkout, view orders, wishlist, and profile."],
            ["Seller", "Manage products, inventory, orders, coupons, campaigns, staff, and analytics."],
            ["Admin", "Manage users, sellers, orders, payments, settings, search, sessions, and audit logs."],
            ["Developer", "Develop services, APIs, frontend modules, tests, and deployments."],
            ["Operations Team", "Monitor health, logs, metrics, alerts, queue lag, and deployment status."],
        ],
        [120, 295],
    )


def add_conclusion(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Conclusion", 0)
    doc.heading("Conclusion", 1)
    doc.paragraph(
        f"The project '{PROJECT_TITLE}' presents a clear design for a scalable and "
        "secure e-commerce platform. It follows a microservices architecture, uses "
        "database-per-service ownership, includes security controls, and plans for "
        "caching, search, message queue, Kubernetes deployment, and observability."
    )
    doc.paragraph(
        "The current implementation proves important parts of the design through a "
        "Go Auth Service and a React TypeScript Session Analytics Dashboard. The Auth "
        "Service demonstrates practical backend development with password security, "
        "OTP, JWT, refresh tokens, RBAC, MySQL, Redis, HTTP routes, and outbox events. "
        "The dashboard demonstrates frontend implementation with filters, typed API "
        "mapping, metric cards, validation, and tests."
    )
    doc.paragraph(
        "This project helped in understanding real software engineering work such as "
        "requirement analysis, modular design, secure coding, database design, API "
        "planning, testing, deployment planning, monitoring, and documentation. The "
        "project is suitable for future expansion into a complete production-ready "
        "e-commerce platform."
    )
    doc.heading("Future Enhancements", 2)
    for item in [
        "Implement remaining services such as User, Product, Cart, Order, Payment, CMS, Search, Recommendation, Notification, and Superadmin.",
        "Add gRPC/protobuf contracts and generated clients.",
        "Complete API Gateway implementation with routing, validation, and rate limiting.",
        "Add full Docker Compose stack and Kubernetes manifests.",
        "Add integration tests and end-to-end tests for checkout and payment flows.",
        "Add charts, journey explorer, funnels, heatmaps, cohorts, and privacy controls in the analytics dashboard.",
        "Add CI/CD pipeline, image scanning, deployment automation, and monitoring dashboards.",
        "Add production-grade secret management and backup/recovery plan.",
    ]:
        doc.bullet(item)


def add_bibliography(doc: ReportDoc) -> None:
    doc.add_page()
    doc.mark("Bibliography (APA Style)", 0)
    doc.heading("Bibliography (APA Style)", 1)
    refs = [
        "Chandigarh University, Centre for Distance & Online Education. (2025). Guidelines for Major Project (23ONMCR-753).",
        "Ecommerce project repository. (2026). README.md, architecture documents, API reference, database design, Auth Service source code, and Session Analytics Dashboard source code.",
        "The Go Authors. (2026). The Go programming language documentation. https://go.dev/doc/",
        "React Team. (2026). React documentation. https://react.dev/",
        "TypeScript Team. (2026). TypeScript documentation. https://www.typescriptlang.org/docs/",
        "MySQL. (2026). MySQL 8.0 reference manual. https://dev.mysql.com/doc/",
        "MongoDB. (2026). MongoDB manual. https://www.mongodb.com/docs/",
        "Redis. (2026). Redis documentation. https://redis.io/docs/",
        "Kubernetes Authors. (2026). Kubernetes documentation. https://kubernetes.io/docs/",
        "OWASP Foundation. (2026). OWASP application security guidance. https://owasp.org/",
        "OpenTelemetry. (2026). OpenTelemetry documentation. https://opentelemetry.io/docs/",
        "Prometheus Authors. (2026). Prometheus documentation. https://prometheus.io/docs/",
        "Grafana Labs. (2026). Grafana and Loki documentation. https://grafana.com/docs/",
        "Typesense. (2026). Typesense documentation. https://typesense.org/docs/",
        "TanStack. (2026). TanStack Query documentation. https://tanstack.com/query/",
    ]
    for ref in refs:
        doc.bullet(ref, size=11.5)


def build_report(toc_pages: dict[str, int] | None = None) -> ReportDoc:
    doc = ReportDoc(toc_pages=toc_pages)
    add_title_page(doc)
    add_certificate(doc)
    add_declaration(doc)
    add_acknowledgement(doc)
    add_abstract(doc)
    doc.add_page()
    if toc_pages is None:
        doc.mark("Table of Contents", 0)
        doc.heading("Table of Contents", 1)
        doc.paragraph("Table of contents will be generated in the final pass.")
    else:
        doc.write_toc()
    add_intro(doc)
    add_sdlc(doc)
    add_design(doc)
    add_implementation(doc)
    add_testing(doc)
    add_application(doc)
    add_conclusion(doc)
    add_bibliography(doc)
    return doc


def make_pdf(doc: ReportDoc, path: Path) -> None:
    objects: dict[int, bytes] = {}
    objects[1] = b"<< /Type /Catalog /Pages 2 0 R >>"
    objects[3] = b"<< /Type /Font /Subtype /Type1 /BaseFont /Times-Roman >>"
    objects[4] = b"<< /Type /Font /Subtype /Type1 /BaseFont /Times-Bold >>"
    objects[5] = b"<< /Type /Font /Subtype /Type1 /BaseFont /Times-Italic >>"
    objects[6] = b"<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>"

    next_obj = 7
    page_ids: list[int] = []
    for page in doc.pages:
        content = ("\n".join(page.ops) + "\n").encode("latin1", "replace")
        content_id = next_obj
        next_obj += 1
        objects[content_id] = b"<< /Length %d >>\nstream\n" % len(content) + content + b"endstream"
        page_id = next_obj
        next_obj += 1
        page_ids.append(page_id)
        objects[page_id] = (
            f"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 {PAGE_W:.2f} {PAGE_H:.2f}] "
            f"/Resources << /Font << /F1 3 0 R /F2 4 0 R /F3 5 0 R /F4 6 0 R >> >> "
            f"/Contents {content_id} 0 R >>"
        ).encode("latin1")
    kids = " ".join(f"{pid} 0 R" for pid in page_ids)
    objects[2] = f"<< /Type /Pages /Count {len(page_ids)} /Kids [ {kids} ] >>".encode("latin1")

    max_obj = max(objects)
    out = bytearray()
    out.extend(b"%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
    offsets = [0] * (max_obj + 1)
    for obj_id in range(1, max_obj + 1):
        offsets[obj_id] = len(out)
        out.extend(f"{obj_id} 0 obj\n".encode("ascii"))
        out.extend(objects[obj_id])
        out.extend(b"\nendobj\n")
    xref = len(out)
    out.extend(f"xref\n0 {max_obj + 1}\n".encode("ascii"))
    out.extend(b"0000000000 65535 f \n")
    for obj_id in range(1, max_obj + 1):
        out.extend(f"{offsets[obj_id]:010d} 00000 n \n".encode("ascii"))
    out.extend(
        f"trailer\n<< /Size {max_obj + 1} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n".encode("ascii")
    )
    path.write_bytes(out)


def main() -> None:
    first = build_report()
    toc = {"entries": first.toc_entries}
    final = build_report(toc)
    make_pdf(final, OUTPUT_PATH)
    print(f"Generated {OUTPUT_PATH} with {len(final.pages)} pages")


if __name__ == "__main__":
    main()
