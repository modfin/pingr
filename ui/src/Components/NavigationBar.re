[@react.component]
let make = () => {
  <nav className="flex items-center justify-between flex-wrap bg-gray-600 p-6">
    <div
      className="flex items-center flex-shrink-0 text-white mr-6 cursor-pointer"
      onClick={_ => Paths.goToTests()}>
      <img
        className="w-8 mb-3"
        src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/62d0c3d5f52dbc9c803ea7bfaa2829d75a2f8fa2.png"
      />
      <span className="font-semibold text-xl tracking-tight ml-1">
        {"Pingr" |> React.string}
      </span>
    </div>
    <div className="w-full block flex-grow md:flex md:items-center md:w-auto">
      <div className="text-md md:flex-grow">
        <a
          onClick={_ => Paths.goToTests()}
          className="cursor-pointer block mt-4 md:inline-block md:mt-0 text-blue-200 hover:text-white mr-4">
          {"Tests" |> React.string}
        </a>
        <a
          onClick={_ => Paths.goToContacts()}
          className="cursor-pointer block mt-4 md:inline-block md:mt-0 text-blue-200 hover:text-white mr-4">
          {"Contacts" |> React.string}
        </a>
        <a
          onClick={_ => Paths.goToIncidents()}
          className="cursor-pointer block mt-4 md:inline-block md:mt-0 text-blue-200 hover:text-white mr-4">
          {"Incidents" |> React.string}
        </a>
      </div>
    </div>
  </nav>;
};