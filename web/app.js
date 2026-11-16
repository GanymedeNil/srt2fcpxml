"use strict";

// ---------- Copy ----------
const STRINGS = {
  zh: {
    lede: "把字幕文件变成 Final Cut Pro 工程，每条字幕都是可以自由修改样式的基本字幕；也能把时间线上的标题导出为 SRT。文件只在你的浏览器里处理，不会上传。",
    repo: "源代码与命令行版",
    tabConvert: "字幕 → FCPXML",
    tabExtract: "FCPXML → 字幕",
    dropSubtitle: "选择或拖入字幕文件",
    dropHint: "SRT、WebVTT、ASS/SSA、TTML、STL",
    bilingual: "双语",
    biNone: "单语",
    biFile: "第二个文件",
    biSplit: "同一文件分行",
    dropSecondary: "选择第二语言的字幕文件",
    splitNote: "每条字幕第一行是主语言，其余行是第二语言。",
    project: "工程",
    fps: "帧率",
    resolution: "分辨率",
    vertical1080: "1080 × 1920 竖屏",
    vertical4k: "2160 × 3840 竖屏",
    custom: "自定义…",
    width: "宽",
    height: "高",
    fpsNote: "帧率要和视频一致，字幕才会对齐到帧。",
    style: "样式",
    fontSize: "字号",
    color: "颜色",
    strokeWidth: "描边宽度",
    strokeColor: "描边颜色",
    positionY: "垂直位置",
    secondarySize: "第二语言字号",
    secondaryColor: "第二语言颜色",
    loading: "正在加载转换器…",
    convert: "生成 FCPXML",
    previewNote: "预览为近似效果，以 Final Cut Pro 中为准。",
    gap: "空隙",
    cues: (n) => `${n} 条字幕`,
    download: (name) => `下载 ${name}`,
    importHint: "在 Final Cut Pro 中选择「文件 → 导入 → XML…」打开它。",
    needFile: "先选择一个字幕文件。",
    needSecondary: "选择第二语言的字幕文件，或改用「单语」。",
    dropFcpxml: "选择或拖入 FCPXML 文件",
    dropFcpxmlHint: ".fcpxml；.fcpxmld 包请选择其中的 Info.fcpxml",
    export: "导出",
    format: "格式",
    include: "包含",
    both: "标题和字幕",
    titles: "仅标题",
    captions: "仅字幕（caption）",
    projectName: "工程名（文件里有多个工程时填写）",
    extract: "导出字幕",
    needFcpxml: "先选择一个 FCPXML 文件。",
    cueListEmpty: "导出的字幕会显示在这里。",
    loadFailed: "转换器加载失败，请刷新页面重试。",
    footer: "srt2fcpxml 在浏览器中运行（WebAssembly），文件不会离开你的电脑。输出 FCPXML 1.7，可导入 Final Cut Pro 10.4.6 及以上版本。",
  },
  en: {
    lede: "Turn subtitle files into a Final Cut Pro project where every subtitle is a Basic Title you can style, or export the titles on a timeline back to SRT. Files are processed in your browser and never uploaded.",
    repo: "Source code and command-line version",
    tabConvert: "Subtitles → FCPXML",
    tabExtract: "FCPXML → Subtitles",
    dropSubtitle: "Choose or drop a subtitle file",
    dropHint: "SRT, WebVTT, ASS/SSA, TTML, STL",
    bilingual: "Bilingual",
    biNone: "One language",
    biFile: "Second file",
    biSplit: "Lines in one file",
    dropSecondary: "Choose the second-language subtitle file",
    splitNote: "The first line of each subtitle is the main language; the other lines are the second language.",
    project: "Project",
    fps: "Frame rate",
    resolution: "Resolution",
    vertical1080: "1080 × 1920 vertical",
    vertical4k: "2160 × 3840 vertical",
    custom: "Custom…",
    width: "Width",
    height: "Height",
    fpsNote: "Use your video's frame rate so subtitles line up with its frames.",
    style: "Style",
    fontSize: "Font size",
    color: "Color",
    strokeWidth: "Outline width",
    strokeColor: "Outline color",
    positionY: "Vertical position",
    secondarySize: "Second language size",
    secondaryColor: "Second language color",
    loading: "Loading the converter…",
    convert: "Make FCPXML",
    previewNote: "Approximate preview; Final Cut Pro has the final word.",
    gap: "Gap",
    cues: (n) => `${n} subtitle${n === 1 ? "" : "s"}`,
    download: (name) => `Download ${name}`,
    importHint: "In Final Cut Pro, choose File → Import → XML… and open it.",
    needFile: "Choose a subtitle file first.",
    needSecondary: "Choose the second-language file, or switch to One language.",
    dropFcpxml: "Choose or drop an FCPXML file",
    dropFcpxmlHint: ".fcpxml; for an .fcpxmld bundle choose the Info.fcpxml inside it",
    export: "Export",
    format: "Format",
    include: "Include",
    both: "Titles and captions",
    titles: "Titles only",
    captions: "Captions only",
    projectName: "Project name (if the file has several)",
    extract: "Export subtitles",
    needFcpxml: "Choose an FCPXML file first.",
    cueListEmpty: "The exported subtitles appear here.",
    loadFailed: "The converter failed to load. Reload the page to try again.",
    footer: "srt2fcpxml runs in your browser (WebAssembly); files never leave your computer. It writes FCPXML 1.7, which Final Cut Pro 10.4.6 and later can import.",
  },
};

const $ = (id) => document.getElementById(id);
const storage = {
  get(k) { try { return localStorage.getItem(k); } catch { return null; } },
  set(k, v) { try { localStorage.setItem(k, v); } catch { /* private mode */ } },
};

let lang = storage.get("lang");
if (!STRINGS[lang]) lang = (navigator.language || "").toLowerCase().startsWith("zh") ? "zh" : "en";
const t = (key, ...args) => {
  const v = STRINGS[lang][key];
  return typeof v === "function" ? v(...args) : v;
};
let ready = false;

function applyLanguage() {
  document.documentElement.lang = lang === "zh" ? "zh-CN" : "en";
  for (const el of document.querySelectorAll("[data-i18n]")) {
    if (el.dataset.i18nDynamic) continue;
    el.textContent = t(el.dataset.i18n);
  }
  for (const a of document.querySelectorAll("a.download")) {
    a.textContent = `↓ ${t("download", a.download)}`;
  }
  $("lang").textContent = lang === "zh" ? "EN" : "中文";
  setReady(ready);
  renderTimelineCount();
}

$("lang").addEventListener("click", () => {
  lang = lang === "zh" ? "en" : "zh";
  storage.set("lang", lang);
  applyLanguage();
});

// ---------- Tabs ----------
for (const tab of document.querySelectorAll('[role="tab"]')) {
  tab.addEventListener("click", () => {
    for (const other of document.querySelectorAll('[role="tab"]')) {
      const on = other === tab;
      other.setAttribute("aria-selected", String(on));
      $(other.getAttribute("aria-controls")).hidden = !on;
    }
  });
}

// ---------- Files ----------
const files = { primary: null, secondary: null, fcpxml: null };

function bindFile(inputId, nameId, key, onChange) {
  const input = $(inputId);
  const drop = input.closest(".drop");
  const show = (file) => {
    files[key] = file;
    const name = $(nameId);
    name.dataset.i18nDynamic = file ? "1" : "";
    name.textContent = file ? file.name : t(name.dataset.i18n);
    drop.classList.toggle("has-file", Boolean(file));
    if (onChange) onChange(file);
  };
  input.addEventListener("change", () => show(input.files[0] || null));
  drop.addEventListener("dragover", (e) => { e.preventDefault(); drop.classList.add("over"); });
  drop.addEventListener("dragleave", () => drop.classList.remove("over"));
  drop.addEventListener("drop", (e) => {
    e.preventDefault();
    drop.classList.remove("over");
    if (e.dataTransfer.files[0]) show(e.dataTransfer.files[0]);
  });
}
bindFile("file-primary", "name-primary", "primary", () => clearResult("convert-result"));
bindFile("file-secondary", "name-secondary", "secondary", () => clearResult("convert-result"));
bindFile("file-fcpxml", "name-fcpxml", "fcpxml", () => clearResult("extract-result"));

const bytes = async (file) => new Uint8Array(await file.arrayBuffer());

// ---------- Options ----------
const bilingualMode = () => document.querySelector('input[name="bilingual"]:checked').value;

function size() {
  const v = $("resolution").value;
  if (v === "custom") return [Number($("width").value), Number($("height").value)];
  const [w, h] = v.split("x").map(Number);
  return [w, h];
}

function styleRequest() {
  const style = {
    fontSize: Number($("font-size").value),
    fontColor: $("font-color").value,
    positionY: Number($("position-y").value),
  };
  const stroke = Number($("stroke-width").value);
  if (stroke > 0) {
    style.strokeWidth = stroke;
    style.strokeColor = $("stroke-color").value;
  }
  if (bilingualMode() !== "none") {
    style.secondary = { fontSize: Number($("secondary-size").value), fontColor: $("secondary-color").value };
  }
  return style;
}

for (const radio of document.querySelectorAll('input[name="bilingual"]')) {
  radio.addEventListener("change", () => {
    const mode = bilingualMode();
    $("drop-secondary").hidden = mode !== "file";
    $("split-note").hidden = mode !== "split";
    $("secondary-style").hidden = mode === "none";
    preview.secondaryLines = null;
    renderPreview();
  });
}
$("resolution").addEventListener("change", () => {
  $("custom-size").hidden = $("resolution").value !== "custom";
  renderPreview();
});
for (const id of ["width", "height", "font-size", "font-color", "stroke-width", "stroke-color", "position-y", "secondary-size", "secondary-color"]) {
  $(id).addEventListener("input", renderPreview);
}

// ---------- Viewer ----------
// Sizes and positions are in the Basic Title template's 1920x1080 coordinates.
const preview = { lines: ["你好，世界"], secondaryLines: ["Hello, world"] };

function renderPreview() {
  const viewer = $("viewer");
  const [w, h] = size();
  if (w > 0 && h > 0) viewer.style.aspectRatio = `${w} / ${h}`;
  viewer.style.width = h > w ? "min(100%, 42vh)" : "100%";
  const vw = viewer.clientWidth, vh = viewer.clientHeight;
  const scale = Math.min(vw / 1920, vh / 1080);

  const caption = $("viewer-text");
  const primary = caption.querySelector(".l1");
  const secondary = caption.querySelector(".l2");
  primary.textContent = preview.lines.join("\n");
  const lines2 = preview.secondaryLines || ["Hello, world"];
  secondary.hidden = bilingualMode() === "none" || lines2.length === 0;
  secondary.textContent = lines2.join("\n");

  const shadow = `${(5 * Math.SQRT1_2 * scale).toFixed(2)}px ${(5 * Math.SQRT1_2 * scale).toFixed(2)}px 0 rgba(0,0,0,0.75)`;
  const stroke = Number($("stroke-width").value);
  for (const [el, sizeId, colorId, weight] of [[primary, "font-size", "font-color", 600], [secondary, "secondary-size", "secondary-color", 400]]) {
    el.style.fontSize = `${Number($(sizeId).value) * scale}px`;
    el.style.color = $(colorId).value;
    el.style.fontWeight = weight;
    el.style.fontFamily = '"PingFang SC", system-ui, sans-serif';
    el.style.textShadow = shadow;
    el.style.webkitTextStroke = stroke > 0 ? `${stroke * scale}px ${$("stroke-color").value}` : "";
  }
  caption.style.top = `${vh / 2 - Number($("position-y").value) * scale}px`;
}
window.addEventListener("resize", renderPreview);

// ---------- Timeline ----------
const timeline = { cues: [], total: 0, selected: -1 };

const seconds = (v) => {
  if (!v) return 0;
  const [n, d] = v.replace(/s$/, "").split("/");
  return Number(n) / (d ? Number(d) : 1);
};
const timecode = (s) => {
  s = Math.max(0, Math.round(s));
  return [Math.floor(s / 3600), Math.floor(s / 60) % 60, s % 60].map((x) => String(x).padStart(2, "0")).join(":");
};

// cueLines splits a title's text into lines, marking the second-language ones,
// which use the second-language font size.
function cueLines(title, secondarySize) {
  const sizes = new Map([...title.querySelectorAll("text-style-def")].map((d) => [d.getAttribute("id"), d.querySelector("text-style")?.getAttribute("fontSize")]));
  const lines = [{ text: "", secondary: null }];
  for (const run of title.querySelectorAll("text > text-style")) {
    const secondary = secondarySize !== null && sizes.get(run.getAttribute("ref")) === String(secondarySize);
    run.textContent.split("\n").forEach((part, i) => {
      if (i > 0) lines.push({ text: "", secondary: null });
      const line = lines[lines.length - 1];
      if (line.secondary === null && part !== "") line.secondary = secondary;
      line.text += part;
    });
  }
  return lines.filter((l) => l.text !== "");
}

function showTimeline(fcpxml, secondarySize) {
  const doc = new DOMParser().parseFromString(fcpxml, "application/xml");
  const gap = doc.querySelector("spine > gap");
  const sequence = doc.querySelector("sequence");
  if (!gap || !sequence) return;
  const base = seconds(gap.getAttribute("start"));
  timeline.total = seconds(sequence.getAttribute("duration"));
  timeline.cues = [...gap.querySelectorAll(":scope > title")].map((title) => {
    const lines = cueLines(title, secondarySize);
    return {
      start: seconds(title.getAttribute("offset")) - base,
      duration: seconds(title.getAttribute("duration")),
      lines,
      text: lines.map((l) => l.text).join("\n"),
    };
  });
  const lane = $("lane");
  lane.replaceChildren();
  timeline.cues.forEach((cue, i) => {
    const clip = document.createElement("button");
    clip.type = "button";
    clip.className = "clip";
    clip.style.left = `${(cue.start / timeline.total) * 100}%`;
    clip.style.width = `${(cue.duration / timeline.total) * 100}%`;
    clip.title = `${timecode(cue.start)}  ${cue.text.replace(/\n/g, " / ")}`;
    clip.setAttribute("aria-label", clip.title);
    clip.addEventListener("click", () => selectCue(i));
    lane.append(clip);
  });
  $("timeline").hidden = false;
  renderTimelineCount();
  selectCue(0);
}

function renderTimelineCount() {
  if (!timeline.cues.length) return;
  $("cue-count").textContent = t("cues", timeline.cues.length);
  $("tc-end").textContent = timecode(timeline.total);
}

function selectCue(i) {
  const cue = timeline.cues[i];
  if (!cue) return;
  timeline.selected = i;
  [...$("lane").children].forEach((c, j) => c.classList.toggle("on", j === i));
  preview.lines = cue.lines.filter((l) => !l.secondary).map((l) => l.text);
  preview.secondaryLines = cue.lines.filter((l) => l.secondary).map((l) => l.text);
  renderPreview();
}

// ---------- Results ----------
function clearResult(id) { $(id).replaceChildren(); }

function showResult(id, response, extraNote) {
  const box = $(id);
  box.replaceChildren();
  if (response.error) {
    const p = document.createElement("p");
    p.className = "error";
    p.textContent = response.error;
    box.append(p);
  }
  if (response.output) {
    const a = document.createElement("a");
    a.className = "download";
    a.href = URL.createObjectURL(new Blob([response.output], { type: "application/octet-stream" }));
    a.download = response.filename;
    a.textContent = `↓ ${t("download", response.filename)}`;
    box.append(a);
    if (extraNote) {
      const p = document.createElement("p");
      p.className = "note";
      p.dataset.i18n = extraNote;
      p.textContent = t(extraNote);
      box.append(p);
    }
  }
  if (response.warnings && response.warnings.length) {
    const ul = document.createElement("ul");
    for (const w of response.warnings) {
      const li = document.createElement("li");
      li.textContent = w;
      ul.append(li);
    }
    box.append(ul);
  }
}

$("convert-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!ready) return;
  if (!files.primary) return showResult("convert-result", { error: t("needFile") });
  const mode = bilingualMode();
  if (mode === "file" && !files.secondary) return showResult("convert-result", { error: t("needSecondary") });
  const [width, height] = size();
  const request = {
    name: files.primary.name,
    secondaryName: mode === "file" ? files.secondary.name : "",
    fps: $("fps").value,
    width, height,
    splitLines: mode === "split",
    style: styleRequest(),
  };
  const secondary = mode === "file" ? await bytes(files.secondary) : null;
  const response = JSON.parse(window.srt2fcpxml.convert(JSON.stringify(request), await bytes(files.primary), secondary));
  showResult("convert-result", response, "importHint");
  // Second-language runs are told apart by their font size, when it differs.
  const sizes = request.style;
  const secondarySize = sizes.secondary && sizes.secondary.fontSize !== sizes.fontSize ? sizes.secondary.fontSize : null;
  if (response.output) showTimeline(response.output, secondarySize);
});

$("extract-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!ready) return;
  if (!files.fcpxml) return showResult("extract-result", { error: t("needFcpxml") });
  const request = { name: files.fcpxml.name, format: $("format").value, only: $("only").value, project: $("project-name").value.trim() };
  const response = JSON.parse(window.srt2fcpxml.extract(JSON.stringify(request), await bytes(files.fcpxml)));
  showResult("extract-result", response);
  const list = $("cue-list");
  list.dataset.i18nDynamic = response.output ? "1" : "";
  list.textContent = response.output ? response.output.replace(/^﻿/, "") : t("cueListEmpty");
});

// ---------- Start ----------
function setReady(on) {
  ready = on;
  $("convert-button").textContent = on ? t("convert") : t("loading");
  $("extract-button").textContent = on ? t("extract") : t("loading");
  $("convert-button").disabled = !on;
  $("extract-button").disabled = !on;
}

applyLanguage();
renderPreview();

window.addEventListener("srt2fcpxml-ready", () => setReady(true));
const go = new Go();
WebAssembly.instantiateStreaming(fetch("srt2fcpxml.wasm"), go.importObject)
  .then((result) => go.run(result.instance))
  .catch((err) => {
    console.error(err);
    $("convert-button").textContent = t("loadFailed");
    $("extract-button").textContent = t("loadFailed");
  });
