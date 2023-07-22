import { PanelPlugin } from "@grafana/data";
import { FlamegraphPanel } from "./components/FlamegraphPanel"
import { Options } from "./types"

export const plugin = new PanelPlugin<Options>(FlamegraphPanel)
