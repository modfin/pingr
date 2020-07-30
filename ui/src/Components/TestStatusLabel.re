[@react.component]
let make = (~active, ~statusId) => {
  switch (active, statusId) {
  | (false, _) =>
    <div className="w-4 h-4 rounded bg-gray-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Paused" |> React.string}
      </span>
    </div>
  | (true, 1) =>
    <div className="w-4 h-4 rounded bg-green-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Success" |> React.string}
      </span>
    </div>
  | (true, 2) =>
    <div className="w-4 h-4 rounded bg-red-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Error" |> React.string}
      </span>
    </div>
  | (true, 3) =>
    <div className="w-4 h-4 rounded bg-red-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Timed out" |> React.string}
      </span>
    </div>
  | (true, 5) =>
    <div className="w-4 h-4 rounded bg-yellow-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Initialized" |> React.string}
      </span>
    </div>
  | _ =>
    <div className="w-4 h-4 rounded bg-gray-500 tooltip">
      <span className="tooltip-text bg-gray-400 border text-black -mt-12">
        {"Unknown status" |> React.string}
      </span>
    </div>
  };
};