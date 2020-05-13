<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher()
  export let soundFile = {};
  export let fgColor = "white";
  export let buttonColors = [];
  const randomIndex = Math.floor(Math.random() * buttonColors.length);
  const bgColor = buttonColors[randomIndex];

   const playSound = () => {
    const playMsg = {
      type: "play",
      soundfiles: [soundFile],
    };
    dispatch("play", playMsg);
  };


</script>

<style>
  button {
    border: none;
    border-radius: 1rem;
    height: 8rem;
    width: 8rem;
    font-weight: bold;
    font-size: 1rem;
    background-color: var(--bgColor);
    color: var(--fgColor);
    transition: background-color 0.2s;
  }

  button:active,
  button:focus {
    border: none;
    outline: none;
    background-color: #2a9d8f;
  }

  button::-moz-focus-inner {
    border: 0;
  }

  @media (max-width: 300px) {
   button {
     height: 6rem;
    width: 6rem;
    font-size: 0.7rem;
   }
}
</style>

<button style="--bgColor: {bgColor}; --fgColor: {fgColor}" on:click={playSound}>
  {soundFile.name}
</button>
