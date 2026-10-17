import type {ExpeditionRun} from "../../../apps/web/src/Expedition";
// The server continues ticking while the browser reloads. No choice or restart
// is sent by this test, so only one automatic phase transition is admissible.
export function continuesExpedition(before:ExpeditionRun,after:ExpeditionRun) {
 if(before.seed_label!==after.seed_label||before.stage!==after.stage||after.seconds_left>before.seconds_left)return false;
 if(["cilia","digestion","cohesion"].some(id=>before.ranks[id]!==after.ranks[id]))return false;
 if(after.revision===before.revision)return after.phase===before.phase;
 return after.revision===before.revision+1&&before.phase==="running"&&
  ["failed","fallen",before.stage===3?"victory":"choice"].includes(after.phase);
}
