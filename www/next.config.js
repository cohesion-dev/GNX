/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  // output: 'export',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
  typescript: {
    ignoreBuildErrors: false,
  },
  async rewrites() {
    return [
      // 路径重写代理
      {
        source: '/api/:path*',
        destination: 'http://127.0.0.1:9000/:path*',
      },
    ]
  },
}

module.exports = nextConfig
