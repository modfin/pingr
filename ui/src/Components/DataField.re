[@react.component]
let make = (~labelName: string, ~value: string) => {
  <div className="md:flex md:items-center mb-6 mx-1">
    <div className="w-20">
      <label
        className="text-md block text-gray-700 font-bold mb-1 md:mb-0 pr-4">
        {labelName |> React.string}
      </label>
    </div>
    <div className="md:w-2/3">
      <input
        className="bg-gray-200 appearance-none border-2 border-gray-200 rounded w-full py-2 px-4 text-gray-700"
        id="inline-full-name"
        type_="text"
        value
        disabled=true
      />
    </div>
  </div>;
};