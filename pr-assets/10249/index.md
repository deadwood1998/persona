# Launcher column review evidence (#10249)

All data is synthetic. Shared panel: 1280/390 × light/dark × standard/large; large profiles also use Reduce Motion. New Agent: four width/theme profiles in zh-Hans, plus desktop light in English. Finite animations are settled for capture.

The reduced-motion baseline retains its opening clip in the captured compositor frame. The new panel explicitly skips that animation; the fixture checks animation-name is none.

## Shared panel before and after

### 1280-light-standard

| Before | After |
| --- | --- |
| ![Before](before-1280-light-standard.png) | ![After](after-1280-light-standard.png) |

[Flat search](search-1280-light-standard.png) · [Twelve launchers](many-1280-light-standard.png)

### 1280-dark-standard

| Before | After |
| --- | --- |
| ![Before](before-1280-dark-standard.png) | ![After](after-1280-dark-standard.png) |

[Flat search](search-1280-dark-standard.png) · [Twelve launchers](many-1280-dark-standard.png)

### 390-light-standard

| Before | After |
| --- | --- |
| ![Before](before-390-light-standard.png) | ![After](after-390-light-standard.png) |

[Flat search](search-390-light-standard.png) · [Twelve launchers](many-390-light-standard.png)

### 390-dark-standard

| Before | After |
| --- | --- |
| ![Before](before-390-dark-standard.png) | ![After](after-390-dark-standard.png) |

[Flat search](search-390-dark-standard.png) · [Twelve launchers](many-390-dark-standard.png)

### 1280-light-large

| Before | After |
| --- | --- |
| ![Before](before-1280-light-large.png) | ![After](after-1280-light-large.png) |

[Flat search](search-1280-light-large.png) · [Twelve launchers](many-1280-light-large.png)

### 1280-dark-large

| Before | After |
| --- | --- |
| ![Before](before-1280-dark-large.png) | ![After](after-1280-dark-large.png) |

[Flat search](search-1280-dark-large.png) · [Twelve launchers](many-1280-dark-large.png)

### 390-light-large

| Before | After |
| --- | --- |
| ![Before](before-390-light-large.png) | ![After](after-390-light-large.png) |

[Flat search](search-390-light-large.png) · [Twelve launchers](many-390-light-large.png)

### 390-dark-large

| Before | After |
| --- | --- |
| ![Before](before-390-dark-large.png) | ![After](after-390-dark-large.png) |

[Flat search](search-390-dark-large.png) · [Twelve launchers](many-390-dark-large.png)

## Full New Agent flows

### default

![web-new-agent-model-default-1280-dark](web-new-agent-model-default-1280-dark.png)

![web-new-agent-model-default-1280-light-en](web-new-agent-model-default-1280-light-en.png)

![web-new-agent-model-default-1280-light](web-new-agent-model-default-1280-light.png)

![web-new-agent-model-default-390-dark](web-new-agent-model-default-390-dark.png)

![web-new-agent-model-default-390-light](web-new-agent-model-default-390-light.png)

### panel

![web-new-agent-model-panel-1280-dark](web-new-agent-model-panel-1280-dark.png)

![web-new-agent-model-panel-1280-light-en](web-new-agent-model-panel-1280-light-en.png)

![web-new-agent-model-panel-1280-light](web-new-agent-model-panel-1280-light.png)

![web-new-agent-model-panel-390-dark](web-new-agent-model-panel-390-dark.png)

![web-new-agent-model-panel-390-light](web-new-agent-model-panel-390-light.png)

### search

![web-new-agent-model-search-1280-dark](web-new-agent-model-search-1280-dark.png)

![web-new-agent-model-search-1280-light-en](web-new-agent-model-search-1280-light-en.png)

![web-new-agent-model-search-1280-light](web-new-agent-model-search-1280-light.png)

![web-new-agent-model-search-390-dark](web-new-agent-model-search-390-dark.png)

![web-new-agent-model-search-390-light](web-new-agent-model-search-390-light.png)

### flyout

![web-new-agent-model-flyout-1280-dark](web-new-agent-model-flyout-1280-dark.png)

![web-new-agent-model-flyout-1280-light-en](web-new-agent-model-flyout-1280-light-en.png)

![web-new-agent-model-flyout-1280-light](web-new-agent-model-flyout-1280-light.png)

![web-new-agent-model-flyout-390-dark](web-new-agent-model-flyout-390-dark.png)

![web-new-agent-model-flyout-390-light](web-new-agent-model-flyout-390-light.png)

### engines

![web-new-agent-model-engines-1280-dark](web-new-agent-model-engines-1280-dark.png)

![web-new-agent-model-engines-1280-light-en](web-new-agent-model-engines-1280-light-en.png)

![web-new-agent-model-engines-1280-light](web-new-agent-model-engines-1280-light.png)

![web-new-agent-model-engines-390-dark](web-new-agent-model-engines-390-dark.png)

![web-new-agent-model-engines-390-light](web-new-agent-model-engines-390-light.png)

### codex

![web-new-agent-model-codex-1280-dark](web-new-agent-model-codex-1280-dark.png)

![web-new-agent-model-codex-1280-light-en](web-new-agent-model-codex-1280-light-en.png)

![web-new-agent-model-codex-1280-light](web-new-agent-model-codex-1280-light.png)

![web-new-agent-model-codex-390-dark](web-new-agent-model-codex-390-dark.png)

![web-new-agent-model-codex-390-light](web-new-agent-model-codex-390-light.png)

### dropped

![web-new-agent-model-dropped-1280-dark](web-new-agent-model-dropped-1280-dark.png)

![web-new-agent-model-dropped-1280-light-en](web-new-agent-model-dropped-1280-light-en.png)

![web-new-agent-model-dropped-1280-light](web-new-agent-model-dropped-1280-light.png)

![web-new-agent-model-dropped-390-dark](web-new-agent-model-dropped-390-dark.png)

![web-new-agent-model-dropped-390-light](web-new-agent-model-dropped-390-light.png)

### follow

![web-new-agent-model-follow-1280-dark](web-new-agent-model-follow-1280-dark.png)

![web-new-agent-model-follow-1280-light-en](web-new-agent-model-follow-1280-light-en.png)

![web-new-agent-model-follow-1280-light](web-new-agent-model-follow-1280-light.png)

![web-new-agent-model-follow-390-dark](web-new-agent-model-follow-390-dark.png)

![web-new-agent-model-follow-390-light](web-new-agent-model-follow-390-light.png)

## Local feedback timing

[Before](before-results.json) · [After](after-results.json). One event-to-next-frame sample per profile; before changes the old provider view, after expands the new launcher. These are local UI feedback measurements, not model execution benchmarks.
