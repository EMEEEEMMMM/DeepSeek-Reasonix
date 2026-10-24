// @vitest-environment jsdom
import { render, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { Wallpaper } from "../port/look";
import { WallpaperScene } from "./WallpaperScene";

const scene: Wallpaper = {
  url: "/bg.jpg",
  kind: "scene",
  title: "花海",
  audio: "/scene.mp3",
  parallax: 0.04,
  muted: false,
  volume: 0.5,
  opacity: 0.5,
  dim: 0.55,
  focusX: 0.5,
  focusY: 0.5,
};

afterEach(() => {
  vi.restoreAllMocks();
  document.documentElement.style.removeProperty("--bg-x");
  document.documentElement.style.removeProperty("--bg-y");
});

describe("a Wallpaper Engine backdrop", () => {
  it("renders a looping soundtrack and starts it", async () => {
    const play = vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
    const { container } = render(<WallpaperScene wallpaper={scene} />);
    const audio = container.querySelector("audio");
    expect(audio?.getAttribute("src")).toBe("/scene.mp3");
    expect(audio?.loop).toBe(true);
    await waitFor(() => expect(audio?.volume).toBeCloseTo(0.5));
    await waitFor(() => expect(play).toHaveBeenCalled());
  });

  it("does not start at zero volume", () => {
    const play = vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
    render(<WallpaperScene wallpaper={{ ...scene, volume: 0 }} />);
    expect(play).not.toHaveBeenCalled();
  });

  it("moves the backdrop with the pointer", () => {
    vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
    render(<WallpaperScene wallpaper={scene} />);
    Object.defineProperty(window, "innerWidth", { configurable: true, value: 1000 });
    Object.defineProperty(window, "innerHeight", { configurable: true, value: 500 });
    window.dispatchEvent(new MouseEvent("pointermove", { clientX: 100, clientY: 100 }));
    expect(document.documentElement.style.getPropertyValue("--bg-x")).not.toBe("50%");
    expect(document.documentElement.style.getPropertyValue("--bg-y")).not.toBe("50%");
  });

  it("does not install the soundtrack when muted", () => {
    const { container } = render(<WallpaperScene wallpaper={{ ...scene, muted: true }} />);
    const audio = container.querySelector("audio");
    expect(audio?.muted).toBe(true);
  });
});
