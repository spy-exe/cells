/** Short synthesized feedback. Audio starts only after an explicit player gesture. */
export class GameSound {
  private context: AudioContext | null = null;
  private lastGain = 0;
  enabled = false;
  async enable() {
    this.context ??= new AudioContext();
    await this.context.resume();
    this.enabled = true;
  }
  play(kind: "gain" | "split" | "eject" | "death") {
    const c = this.context;
    if (!this.enabled || !c || c.state !== "running") return;
    if (kind === "gain" && performance.now() - this.lastGain < 240) return;
    if (kind === "gain") this.lastGain = performance.now();
    const frequencies = { gain: [520, 740], split: [310, 470], eject: [230, 320], death: [190, 80] };
    const [from, to] = frequencies[kind], oscillator = c.createOscillator(), volume = c.createGain();
    const at = c.currentTime, length = kind === "death" ? .24 : .085;
    oscillator.type = "sine";
    oscillator.frequency.setValueAtTime(from, at);
    oscillator.frequency.exponentialRampToValueAtTime(to, at + length);
    volume.gain.setValueAtTime(.0001, at);
    volume.gain.exponentialRampToValueAtTime(.035, at + .01);
    volume.gain.exponentialRampToValueAtTime(.0001, at + length);
    oscillator.connect(volume).connect(c.destination);
    oscillator.start(at); oscillator.stop(at + length);
    oscillator.onended = () => { oscillator.disconnect(); volume.disconnect(); };
  }
  destroy() { void this.context?.close(); this.context = null; }
}
