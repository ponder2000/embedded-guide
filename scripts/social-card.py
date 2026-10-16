"""Regenerate the committed sharing image (optional: Python 3 + Pillow)."""
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
FONT = next(
    (p for p in (
        Path("/System/Library/Fonts/Supplemental/Arial Bold.ttf"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"),
    ) if p.exists()),
    None,
)
if FONT is None:
    raise SystemExit("Install Arial or DejaVu Sans, or set FONT in this script.")

image = Image.new("RGB", (1200, 630), "#fbf8f2")
draw = ImageDraw.Draw(image)
ink, accent, rule = "#292720", "#b8521e", "#d8d1c6"


def text(position, value, size, color=ink):
    draw.text(position, value, font=ImageFont.truetype(str(FONT), size), fill=color)


draw.rectangle((0, 0, 1200, 12), fill=accent)
text((72, 58), "EMBEDDED.GUIDE", 24, accent)
draw.line((72, 110, 1128, 110), fill=rule, width=2)
text((72, 160), "Embedded systems.", 68)
text((72, 252), "A visual guide.", 68, accent)
text((76, 370), "From bits to a blinking chip.", 30)
text((76, 423), "Diagrams. Interactive experiments. Real firmware.", 24)
draw.line((72, 524, 1128, 524), fill=rule, width=2)
text((76, 554), "LEARN BY SEEING", 20, accent)
image.save(ROOT / "web/static/social-card.png", optimize=True)
