<script>
  import SoundButton from "./SoundButton.svelte";
  import VolumeKnob from "./VolumeKnob.svelte";
  import { NotificationDisplay, notifier } from "@beyonk/svelte-notifications";

  import { onMount } from "svelte";

  let volume = 0;
  let sounds = [];
  let connected = false;
  let websocketUrl = location.host;
  let ws;

  const buttonColors = ["#e9c46a", "#f4a261", "#e76f51"];

  const connect = () => {
    const websocket = new WebSocket("ws://" + websocketUrl + "/websocket");

    websocket.onmessage = function(event) {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case "load":
          sounds = data.soundfiles;
          break;
        case "error":
          notifier.danger(data.msg, 5000);
          console.error(data.msg);
          break;
        case "volume":
          volume = data.volume
          break;
        default:
          break;
      }
    };

    websocket.onopen = function(event) {
      ws = websocket;
      connected = true;
    };

    websocket.onclose = function(event) {
      console.info("Connection closed");
      connected = false;
      setTimeout(() => {
        console.info("reconnecting");
        connect();
      }, 5000);
    };

    websocket.onerror = function(event) {
      console.error(event);
      connected = false;
      websocket.close();
      setTimeout(() => {
        console.info("reconnecting");
        connect();
      }, 5000);
    };
  };

  const loadSounds = () => {
    const data = { type: "load" };
    ws.send(JSON.stringify(data));
  };

  const stopSounds = () => {
    const data = { type: "stop" };
    ws.send(JSON.stringify(data));
  };

  const changeVolume = () => {
    const data = { type: "volume", volume: volume };
    ws.send(JSON.stringify(data));
  };

  const playSound = (msg) => {
    ws.send(JSON.stringify(msg))
  }

  onMount(() => {
    connect();
  });
</script>

<style>
  main {
    width: 60vw;
    color: #264653;
    margin: auto;
    text-align: center;
    background-color: white;
  }
  @media (max-width: 1000px) {
    main {
      width: 100vw;
    }
  }
  @media (max-width: 300px) {
    .buttons {
      grid-template-columns: repeat(auto-fill, minmax(4rem, 1fr));
      grid-gap: 0.3rem;
    }
  }

  small {
    color: #999;
  }

  .buttons {
    padding: 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
    grid-gap: 0.5rem;
  }

  .ctrl {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-gap: 1rem;
    padding: 1rem;
  }

  button {
    border: none;
    color: white;
    font-weight: bold;
    background-color: #264653;
    height: 4rem;
    width: 100%;
    border-radius: 1rem;
  }
</style>

<NotificationDisplay />

<main>
  <h1>Go Soundboard</h1>

  <p>
    {#if connected}
      <svg height="10" width="10">
        <circle cx="5" cy="5" r="5" fill="green" />
      </svg>
      online
    {:else}
      <svg height="10" width="10">
        <circle cx="5" cy="5" r="5" fill="red" />
      </svg>
      offline
    {/if}
    <small>{websocketUrl}</small>
  </p>
  <VolumeKnob
    on:valueChanged={changeVolume}
    width="200"
    bind:value={volume}
    height="200"
    knobColor="#264653"
    indicatorColor="#2a9d8f"
    volumeColor="#2a9d8f" />

  <section class="ctrl">
    <button on:click={stopSounds}>STOP</button>
    <button on:click={loadSounds}>RELOAD FILES</button>

  </section>

  <section class="buttons">
    {#each sounds as sound}
      <SoundButton on:play={(e) => playSound(e.detail)} soundFile={sound} {buttonColors} />
    {/each}
  </section>
</main>
