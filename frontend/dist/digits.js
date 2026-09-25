// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The seven-segment face, as SVG.
//
// The geometry is transcribed from the Fyne build's ui/digit_generator.go,
// which drew each digit into a 100x200 raster with github.com/fogleman/gg:
// round-capped lines 15 units wide, at the coordinates below. An SVG line with
// stroke-linecap="round" and stroke-width="15" is the same shape, so the face
// keeps its proportions without anyone having to redesign it.
//
// The separator was a 45x200 cell with two filled circles of radius 15.

const DIGIT_WIDTH = 100;
const SEPARATOR_WIDTH = 45;
const CELL_HEIGHT = 200;
const STROKE_WIDTH = 15;

// Segments A-G, in the order the Go code numbered them 0-6.
const SEGMENTS = [
  { x1: 19.5, y1: 12.5, x2: 80.5, y2: 12.5 },  // A - top
  { x1: 92.5, y1: 29.5, x2: 92.5, y2: 90.5 },  // B - upper right
  { x1: 92.5, y1: 109.5, x2: 92.5, y2: 170.5 }, // C - lower right
  { x1: 19.5, y1: 187.5, x2: 80.5, y2: 187.5 }, // D - bottom
  { x1: 7.5, y1: 109.5, x2: 7.5, y2: 170.5 },  // E - lower left
  { x1: 7.5, y1: 29.5, x2: 7.5, y2: 90.5 },    // F - upper left
  { x1: 22.5, y1: 100, x2: 77.5, y2: 100 },    // G - middle
];

// Which segments each digit lights, copied from the Go table.
const DIGIT_SEGMENTS = [
  [0, 1, 2, 3, 4, 5],    // 0
  [1, 2],                // 1
  [0, 1, 6, 4, 3],       // 2
  [0, 1, 6, 2, 3],       // 3
  [5, 6, 1, 2],          // 4
  [0, 5, 6, 2, 3],       // 5
  [0, 5, 4, 3, 2, 6],    // 6
  [0, 1, 2],             // 7
  [0, 1, 2, 3, 4, 5, 6], // 8
  [0, 1, 2, 3, 5, 6],    // 9
];

const SVG_NS = "http://www.w3.org/2000/svg";

function el(name, attrs) {
  const node = document.createElementNS(SVG_NS, name);
  for (const [key, value] of Object.entries(attrs)) {
    node.setAttribute(key, String(value));
  }
  return node;
}

/**
 * Build the face and return a handle for updating it.
 *
 * The structure is built once per layout change rather than once per second:
 * a tick only flips segment classes, so the browser has nothing to lay out
 * again. This is the whole reason the tick moved to the frontend.
 *
 * @param {boolean} showSeconds whether the face carries a seconds pair.
 * @returns {{svg: SVGSVGElement, setTime: (digits: string) => void}}
 */
export function buildFace(showSeconds) {
  const digitCount = showSeconds ? 6 : 4;
  // HH : MM ( : SS ), so a separator after every pair but the last.
  const separatorCount = showSeconds ? 2 : 1;
  const width = digitCount * DIGIT_WIDTH + separatorCount * SEPARATOR_WIDTH;

  const svg = el("svg", {
    viewBox: `0 0 ${width} ${CELL_HEIGHT}`,
    class: "face-svg",
    role: "img",
    "aria-label": "clock face",
  });

  // segmentsByDigit[i][s] is digit i's segment s, so setTime can address one
  // line without searching the DOM.
  const segmentsByDigit = [];
  let x = 0;

  for (let i = 0; i < digitCount; i++) {
    const group = el("g", { transform: `translate(${x}, 0)` });
    const segments = SEGMENTS.map((s) =>
      el("line", {
        x1: s.x1,
        y1: s.y1,
        x2: s.x2,
        y2: s.y2,
        class: "segment",
        "stroke-width": STROKE_WIDTH,
        "stroke-linecap": "round",
      })
    );
    for (const segment of segments) group.appendChild(segment);
    svg.appendChild(group);
    segmentsByDigit.push(segments);
    x += DIGIT_WIDTH;

    // After the second and, with seconds shown, the fourth digit.
    if (i % 2 === 1 && i < digitCount - 1) {
      const colon = el("g", { transform: `translate(${x}, 0)`, class: "colon" });
      colon.appendChild(el("circle", { cx: 22.5, cy: 60, r: 15 }));
      colon.appendChild(el("circle", { cx: 22.5, cy: 140, r: 15 }));
      svg.appendChild(colon);
      x += SEPARATOR_WIDTH;
    }
  }

  /**
   * Light the segments for a run of digit characters, e.g. "0930".
   * Characters past the face's width, or that are not digits, are ignored.
   */
  function setTime(digits) {
    for (let i = 0; i < segmentsByDigit.length; i++) {
      const value = Number(digits[i]);
      const lit = Number.isInteger(value) ? DIGIT_SEGMENTS[value] : [];
      const segments = segmentsByDigit[i];
      for (let s = 0; s < segments.length; s++) {
        segments[s].classList.toggle("on", lit.includes(s));
      }
    }
  }

  return { svg, setTime };
}
