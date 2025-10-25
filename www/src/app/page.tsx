import { observer } from 'mobx-react-lite'

const Home = observer(() => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="max-w-md w-full mx-4 p-8 bg-white rounded-2xl shadow-xl">
        <h1 className="text-3xl font-bold text-gray-800 mb-4 text-center">
          Welcome to GNX
        </h1>
        <p className="text-gray-600 text-center mb-6">
          Next.js + TypeScript + MobX + Tailwind CSS
        </p>
        <div className="space-y-2 text-sm text-gray-500">
          <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <span>CSR Mode</span>
            <span className="text-green-600 font-semibold">✓</span>
          </div>
          <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <span>Mobile Responsive</span>
            <span className="text-green-600 font-semibold">✓</span>
          </div>
        </div>
      </div>
    </div>
  )
})

export default Home
