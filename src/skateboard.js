const { Component, useState } = React;
const { to, animated, useSpring } = ReactSpring;
const { useDrag } = ReactUseGesture;
const PI = Math.PI;
const finalistColor = {
  "0": "#f89eab",
  "-400": "#ceb7ff",
  "-800": "#9dd1fc",
  "-1200": "#affdff",
};
const lerp = (minX, maxX, minY, maxY, clampFlag) => {
  var slope = (maxY - minY) / (maxX - minX);
  return clampFlag
    ? function (x) {
        return ((x < minX ? minX : x > maxX ? maxX : x) - minX) * slope + minY;
      }
    : function (x) {
        return (x - minX) * slope + minY;
      };
};
const scaleRange = [
  -1200,
  -1000,
  -800,
  -600,
  -400,
  -200,
  0,
  200,
  400,
  600,
  800,
  1000,
  1200,
];
const scaleOutput = [1, 1.4, 1, 1.4, 1, 1.4, 1, 1.4, 1, 1.4, 1, 1.4, 1];
const myScale = d3.scaleLinear().domain(scaleRange).range(scaleOutput);
const deckModel =
  "https://raw.githubusercontent.com/pizza3/asset/master/untitled.obj";
const deckMaterials = [
  "https://raw.githubusercontent.com/pizza3/asset/master/mat1.mtl",
  "https://raw.githubusercontent.com/pizza3/asset/master/mat2.mtl",
  "https://raw.githubusercontent.com/pizza3/asset/master/mat3.mtl",
  "https://raw.githubusercontent.com/pizza3/asset/master/mat4.mtl",
];
const wheelModel =
  "https://raw.githubusercontent.com/pizza3/asset/master/wheel.obj";
const wheelMaterials =
  "https://raw.githubusercontent.com/pizza3/asset/master/wheel.mtl";
const iconSet = {
  menu: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="3" y1="12" x2="21" y2="12"></line>
      <line x1="3" y1="6" x2="21" y2="6"></line>
      <line x1="3" y1="18" x2="21" y2="18"></line>
    </svg>
  ),
  cross: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="18" y1="6" x2="6" y2="18"></line>
      <line x1="6" y1="6" x2="18" y2="18"></line>
    </svg>
  ),
  left: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="19" y1="12" x2="5" y2="12"></line>
      <polyline points="12 19 5 12 12 5"></polyline>
    </svg>
  ),
  right: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="5" y1="12" x2="19" y2="12"></line>
      <polyline points="12 5 19 12 12 19"></polyline>
    </svg>
  ),
  insta: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="2" y="2" width="20" height="20" rx="5" ry="5"></rect>
      <path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z"></path>
      <line x1="17.5" y1="6.5" x2="17.51" y2="6.5"></line>
    </svg>
  ),
  twit: (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M23 3a10.9 10.9 0 0 1-3.14 1.53 4.48 4.48 0 0 0-7.86 3v1A10.66 10.66 0 0 1 3 4s-4 9 5 13a11.64 11.64 0 0 1-7 2c9 5 20 0 20-11.5a4.5 4.5 0 0 0-.08-.83A7.72 7.72 0 0 0 23 3z"></path>
    </svg>
  ),
};
class App2 extends Component {
  deckArray = [undefined, undefined, undefined, undefined];

  componentDidMount() {
    this.initRender();
    this.addLighting();
    this.addModel();
    this.addWheels();
    this.animateRender();
    window.addEventListener("resize", this.handleResize, false);
  }
  componentWillUnmount() {
    window.removeEventListener("resize", this.handleResize, false);
  }

  componentDidUpdate(prevProps) {
    if (this.props.active !== prevProps.active) {
      this.deckArray[Math.abs(prevProps.active)].visible = false;
      this.deckArray[Math.abs(this.props.active)].visible = true;
    }
  }

  initRender() {
    this.scene = new THREE.Scene();
    this.camera = new THREE.PerspectiveCamera(
      75,
      400 / window.innerHeight,
      0.1,
      1000
    );
    this.camera.position.z = 5;

    this.renderer = new THREE.WebGLRenderer({ alpha: true });
    this.renderer.shadowMap.enabled = true;
    this.renderer.antialias = true;
    this.renderer.setPixelRatio(window.devicePixelRatio);
    this.renderer.setSize(400, window.innerHeight);
    this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    this.renderer.interpolateneMapping = THREE.ACESFilmicToneMapping;
    this.renderer.outputEncoding = THREE.sRGBEncoding;
    this.renderer.setClearAlpha(0);
    this.group = new THREE.Group();
    this.scene.add(this.group);
    document.getElementById("world").appendChild(this.renderer.domElement);
    this.renderer.render(this.scene, this.camera);
  }

  addLighting = () => {
    const hemislight = new THREE.HemisphereLight();
    hemislight.intensity = 0.2;
    this.scene.add(hemislight);
    const pointlight = new THREE.PointLight();
    pointlight.distance = 1000;
    pointlight.intensity = 0.7;
    pointlight.position.set(30, 70, 20);
    this.scene.add(pointlight);
  };

  addModel = () => {
    deckMaterials.forEach((data, index) => {
      this.addDeck(data, index);
    });
  };

  addDeck = (data, index) => {
    const loader = new THREE.OBJLoader();
    const mtlloader = new THREE.MTLLoader();
    const Scene = this.group;
    const DeckArray = this.deckArray;
    const active = this.props.active;
    mtlloader.load(data, (materials) => {
      materials.preload();
      loader.setMaterials(materials);
      loader.load(deckModel, function (object) {
        object.scale.set(0.5, 0.5, 0.5);
        object.position.set(0, 0.3, 0);
        object.rotation.set(PI / 2, PI, PI);
        if (active !== index) {
          object.visible = false;					 
        }
        DeckArray[index] = object;
        Scene.add(object);
      });
    });
  };

  addWheels = () => {
    const loader = new THREE.OBJLoader();
    const mtlloader = new THREE.MTLLoader();
    const Scene = this.group;
    mtlloader.load(wheelMaterials, (materials) => {
      materials.preload();
      loader.setMaterials(materials);
      loader.load(wheelModel, function (object) {
        object.scale.set(0.5, 0.5, 0.5);
        object.position.set(0, 0.3, 0);
        object.rotation.set(PI / 2, PI, PI);					 
        Scene.add(object);
      });
    });
  };

  animateRender() {
    this.renderer.render(this.scene, this.camera);
    this.group.rotation.set(0, lerp(0, 1200, 0, 6 * PI)(this.props.anim), 0);
    const sVal = myScale(this.props.anim);
    this.group.scale.set(sVal, sVal, sVal);
    this._frameId = window.requestAnimationFrame(this.animateRender.bind(this));
  }
  handleResize = () => {
    this.camera.aspect = 400 / window.innerHeight;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(400, window.innerHeight);
  };

  render() {
    return <div id="world" className="worldContainer"></div>;
  }
}

const App = () => {
  const [animRot, setAnimRot] = useState(0);
  const [isStat, setStat] = useState(false);
  const [active, setActive] = useState(0);
  const [{ trans }, setTrans] = useSpring(() => ({
    trans: 0,
  }));
  const [{ rot2 }, set2] = useSpring(() => ({
    rot2: 0,
    onFrame: ({ rot2 }) => {
      setAnimRot(rot2);
      handleActive(rot2);
    },
  }));
  const [{ rot3 }, set3] = useSpring(() => ({
    rot3: 0,
    config: { mass: 0.5 },
  }));

  const handleActive = (rot2) => {
    if (rot2 >= -200) {
      setActive(0);
    } else if (rot2 < -200 && rot2 > -600) {
      if (active !== 1) setActive(1);
    } else if (rot2 <= -600 && rot2 > -1000) {
      if (active !== 2) setActive(2);
    } else if (rot2 <= -1000) {
      if (active !== 3) setActive(3);
    }
  };

  const [pos, setPos] = useState(0);
  const bind = useDrag(
    ({ down, movement: [x, y], direction: [xDir], velocity }) => {
      const trigger = velocity > 0.2;
      const dir = xDir < 0 ? -1 : 1;
      if (!isStat) {
        if (trigger && !down) {
          const newPosition = pos + dir * 400;
          if (newPosition !== 400 && newPosition !== -1600) {
            set2({ rot2: newPosition });
            set3({ rot3: newPosition });
            setPos(newPosition);
          } else {
            set2({ rot2: pos });
            set3({ rot3: pos });
          }
        } else {
          set2({ rot2: down ? x + pos : pos });
        }
      }
    }
  );
  const move = (dir) => {
    if (!isStat) {
      const newPosition = pos + dir * 400;
      if (newPosition <= 0 && newPosition >= -1200) {
        set2({ rot2: newPosition });
        set3({ rot3: newPosition });
        setPos(newPosition);
      }
    }
  };
  const handleStat = (bool) => {
    setStat(bool);
    if (!isStat) {
      setTrans({ trans: -196 });
    } else {
      setTrans({ trans: 0 });
    }
  };
  return (
    <div className="App" {...bind()}>
      <animated.div
        style={{
          transform: rot3.interpolate((x) => `translateX(${x}px)`),
        }}
        id="textContainer"
      >
        <div id="text">
          <div id="firstName">TOM</div>
        </div>
        <div id="text">
          <div id="firstName">TONY</div>
        </div>
        <div id="text">
          <div id="firstName">DANNY</div>
        </div>
        <div id="text">
          <div id="firstName">BUCKY</div>
        </div>
      </animated.div>
      <animated.div
        style={{
          top: "20vh",
          transform: rot2.interpolate((x) => `translateX(${x}px)`),
        }}
        id="textContainer"
      >
        <div id="text">
          <div id="lastName">ASTA</div>
        </div>
        <div id="text">
          <div id="lastName">HAWK</div>
        </div>
        <div id="text">
          <div id="lastName">WAY</div>
        </div>
        <div id="text">
          <div id="lastName">LASEK</div>
        </div>
      </animated.div>
      <animated.div
        id="colorWave"
        style={{
          transform: trans
            .interpolate([0, -196], [15, 0])
            .interpolate((x) => `translateY(${x}vh)`),
        }}
      >
        <animated.div id="finalistcontainer">
          <div
            id="finalist"
            style={{ color: finalistColor[pos], opacity: isStat ? 0 : 1 }}
          >
            FINALIST
          </div>
        </animated.div>
        <animated.div
          style={{
            display: "flex",
            transform: rot2.interpolate((x) => `translateX(${x}px)`),
          }}
        >
          <PlayerInfo trans={trans} handleStat={handleStat} />
        </animated.div>
        <animated.div
          id="infoContainer"
          style={{
            opacity: trans.interpolate([0, -196], [0, 1]),
          }}
        >
          <animated.button
            id="cross"
            style={{
              opacity: trans.interpolate([0, -196], [0, 1]),
            }}
            onClick={() => {
              handleStat(false);
            }}
          >
            {iconSet.cross}
          </animated.button>
          <div id="info">AGE</div>
          <div id="data">21</div>
          <div id="info">COUNTRY</div>
          <div id="data">UNITED STATES</div>
          <div id="info">HIGHEST RANK</div>
          <div id="data">WINNER</div>
          <div id="info">STANCE</div>
          <div id="data">REGULAR</div>
          <div id="info">SOCIAL LINKS</div>
          <div id="data">
            <div id="socials">{iconSet.insta}</div>
            <div id="socials">{iconSet.twit}</div>
          </div>
        </animated.div>
      </animated.div>
      <animated.div
        style={{
          transform: trans.interpolate((x) => `translate(${x}px)`),
        }}
      >
        <App2 anim={animRot} active={active} />
      </animated.div>
      <animated.button
        id="profile"
        onClick={() => {
          handleStat(true);
        }}
        style={{
          opacity: trans.interpolate([0, -196], [1, 0]),
        }}
      >
        VIEW STATS
      </animated.button>
      <animated.button
        id="left"
        style={{
          opacity: trans.interpolate([0, -196], [1, 0]),
        }}
        onClick={() => {
          move(1);
        }}
        disabled={active === 0}
      >
        {iconSet.left}
      </animated.button>
      <animated.button
        id="right"
        style={{
          opacity: trans.interpolate([0, -196], [1, 0]),
        }}
        onClick={() => {
          move(-1);
        }}
        disabled={active === 3}
      >
        {iconSet.right}
      </animated.button>
      <button id="menu">{iconSet.menu}</button>
    </div>
  );
};

const PlayerInfo = ({ trans, handleStat }) => {
  return (
    <animated.div
      style={{
        display: "flex",
      }}
    >
      <div id="colorWaveDiv" style={{ backgroundColor: "#403d3d" }}>
        <div id="container">
          <animated.div
            id="containerName"
            style={{
              opacity: trans.interpolate([0, -196], [1, 0]),
            }}
          >
            TOM ASTA
          </animated.div>
        </div>
      </div>
      <div id="colorWaveDiv" style={{ backgroundColor: "#603c5a" }}>
        <div id="container">
          <animated.div
            id="containerName"
            style={{
              opacity: trans.interpolate([0, -196], [1, 0]),
            }}
          >
            TONY HAWK
          </animated.div>
        </div>
      </div>
      <div id="colorWaveDiv" style={{ backgroundColor: "#1e2f3c" }}>
        <div id="container">
          <animated.div
            id="containerName"
            style={{
              opacity: trans.interpolate([0, -196], [1, 0]),
            }}
          >
            DANNY WAY
          </animated.div>
        </div>
      </div>
      <div id="colorWaveDiv" style={{ backgroundColor: "#2d4242" }}>
        <div id="container">
          <animated.div
            id="containerName"
            style={{
              opacity: trans.interpolate([0, -196], [1, 0]),
            }}
          >
            BUCKY LASEK
          </animated.div>
        </div>
      </div>
    </animated.div>
  );
};

ReactDOM.render(<App2 />, document.getElementById("root"));
